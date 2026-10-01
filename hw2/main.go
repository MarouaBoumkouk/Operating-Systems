package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// sink is used by the small work loop in the critical section,
// so the compiler doesn't remove the loop
var sink int

// --------------------
// Ticket Lock (OSTEP Figure 28.7)
// --------------------

// ticket = next ticket number to hand out
// turn   = ticket number that is allowed to enter now
type TicketLock struct {
	ticket uint32
	turn   uint32
}

func (l *TicketLock) Lock() {
	// Take a ticket (like FetchAndAdd).
	// AddUint32 returns the new value, so subtract 1 to get my ticket.
	my := atomic.AddUint32(&l.ticket, 1) - 1

	// Wait until it's my turn
	for atomic.LoadUint32(&l.turn) != my {
		runtime.Gosched() // yield so other goroutines can run
	}
}

func (l *TicketLock) Unlock() {
	// Let the next ticket in (first come, first served)
	atomic.AddUint32(&l.turn, 1)
}

// --------------------
// Compare-and-Swap Spin Lock (OSTEP Ch. 28.9)
// --------------------

// flag = 0 means free, 1 means taken
type CASLock struct {
	flag int32
}

func (l *CASLock) Lock() {
	// Try to change flag from 0 to 1 in one atomic step.
	// If it fails, someone else has the lock, so try again.
	for !atomic.CompareAndSwapInt32(&l.flag, 0, 1) {
		runtime.Gosched() // yield so other goroutines can run
	}
}

func (l *CASLock) Unlock() {
	// Set flag back to 0 so another goroutine can take the lock
	atomic.StoreInt32(&l.flag, 0)
}

// Locker lets both lock types use the same benchmark
type Locker interface {
	Lock()
	Unlock()
}

const (
	iters  = 10000 // lock acquisitions per goroutine
	trials = 5     // number of times to repeat each test
)

// bench runs one lock with n goroutines and returns:
// avg = average waiting time (ns)
// p99 = 99th percentile waiting time (ns), shows the worst waits
// ok  = true if the lock kept mutual exclusion
func bench(newLock func() Locker, n int) (avg, p99 float64, ok bool) {
	ok = true

	for t := 0; t < trials; t++ {
		lk := newLock()                 // new lock for each trial
		counter := 0                    // shared counter protected by the lock
		waits := make([]int64, n*iters) // stores every waiting time
		var wg sync.WaitGroup           // waits for all goroutines to finish

		// Channel used to start all goroutines at the same time
		startSignal := make(chan struct{})

		for g := 0; g < n; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				<-startSignal // wait until all goroutines are created

				for i := 0; i < iters; i++ {
					start := time.Now() // start trying to get the lock
					lk.Lock()
					// Waiting time = time from trying to getting the lock
					waits[g*iters+i] = time.Since(start).Nanoseconds()

					// Critical section: only one goroutine at a time
					counter++
					for k := 0; k < 100; k++ { // small work while holding the lock
						sink += k
					}

					lk.Unlock()
				}
			}(g)
		}

		close(startSignal) // start all goroutines together
		wg.Wait()          // wait for all goroutines to finish

		// If the lock works, counter should equal n * iters.
		// If not, two goroutines were in the critical section at once.
		if counter != n*iters {
			ok = false
		}

		// Sort waiting times so we can find p99
		sort.Slice(waits, func(i, j int) bool { return waits[i] < waits[j] })

		// Add up all waiting times to get the average
		var sum int64
		for _, w := range waits {
			sum += w
		}

		// Average the results across all trials
		avg += float64(sum) / float64(len(waits)) / float64(trials)
		p99 += float64(waits[len(waits)*99/100]) / float64(trials)
	}
	return
}

func main() {
	fmt.Println("lock    goroutines  avg_ns   p99_ns  correct")

	// More goroutines = more contention for the lock
	for _, n := range []int{1, 2, 4, 8, 16, 32} {
		// Ticket lock
		a, p, ok := bench(func() Locker { return &TicketLock{} }, n)
		fmt.Printf("ticket  %-10d  %-7.0f  %-7.0f %v\n", n, a, p, ok)

		// CAS lock (same workload and setup)
		a, p, ok = bench(func() Locker { return &CASLock{} }, n)
		fmt.Printf("cas     %-10d  %-7.0f  %-7.0f %v\n", n, a, p, ok)
	}
}
