# EECE 4811/5811 HW2

## Group Members
* Maroua Boumkouk
* Oluchukwu Okeke

## Lock Analysis (2pt)

No, the new algorithm is not correct.

Without the guard, checking `flag` and setting `flag = 1` are no longer protected as one atomic operation. Two threads can both see `flag == 0` before either one changes it.

Example execution:

1. `flag = 0`
2. Thread 1 checks `flag` and sees `0`.
3. The scheduler switches to Thread 2 before Thread 1 sets it to `1`.
4. Thread 2 also checks `flag` and sees `0`.
5. Thread 2 sets `flag = 1` and enters the critical section.
6. The scheduler switches back to Thread 1.
7. Thread 1 also sets `flag = 1` and enters the critical section.

Now both threads are in the critical section at the same time.

Therefore, removing the guard breaks **mutual exclusion**, so the new lock algorithm is incorrect.

## Lock Analysis II (1pt)

No, the modified lock is not correct.

`lock->flag = lock->flag - 1` is not atomic. It reads `flag`, subtracts 1, and then writes the new value.

A bad execution can happen if a thread calls `unlock()` without holding the lock:

1. `flag = 0`
2. T2 calls `unlock()` and reads `flag = 0`, but gets interrupted before writing.
3. T1 calls `lock()`, changes `flag` to `1`, and enters the critical section.
4. T2 continues and writes `-1` to `flag`.
5. T3 calls `lock()`. Since `flag` is not `1`, T3 can acquire the lock and enter the critical section.

Now T1 and T3 are both in the critical section.

So the modified algorithm is incorrect because it breaks **mutual exclusion**.

## Lock Analysis III (1pt)

Without `setpark()`, a thread can miss its wakeup.

Example:

1. T1 holds the lock.
2. T2 sees the lock is busy, adds itself to the queue, and releases the guard.
3. Before T2 calls `park()`, it gets interrupted.
4. T1 calls `unlock()` and `unpark(T2)`.
5. Since T2 is not sleeping yet, the wakeup is lost.
6. T2 resumes and calls `park()`, so it goes to sleep.

Now T2 may sleep forever because its wakeup already happened.

`setpark()` fixes this by telling the OS that T2 is about to sleep. If `unpark(T2)` happens before `park()`, the wakeup is remembered, so `park()` returns immediately.

So `setpark()` prevents the **wakeup/waiting race**.

## Comparing Lock Algorithms (2pt)

### Implementation

Both locks are written in Go in `main.go`, using only `sync/atomic`. No built-in mutex is used.

**Ticket Lock (Figure 28.7)**
- `ticket` is the next ticket to hand out, and `turn` is the ticket allowed in now.
- `Lock()` takes a ticket with `atomic.AddUint32` (like FetchAndAdd), then waits until `turn` equals its ticket.
- `Unlock()` adds 1 to `turn`, so the next goroutine in line gets in.
- Goroutines get the lock in the order they asked (FIFO), so it is fair.

**CAS Spin Lock (Ch. 28.9)**
- `flag` is 0 when free and 1 when taken.
- `Lock()` keeps trying `atomic.CompareAndSwapInt32(&flag, 0, 1)` until it succeeds.
- `Unlock()` sets `flag` back to 0 with `atomic.StoreInt32`.
- Whoever wins the CAS gets the lock, so it is not fair.

Both locks call `runtime.Gosched()` while waiting, so a spinning goroutine yields and lets the lock holder run.

### How to Compile and Run

Requires Go.

```bash
cd hw2
go run main.go
```

### Benchmark Design

- **Contention:** we vary the number of goroutines: 1, 2, 4, 8, 16, 32. More goroutines means more fighting over the lock.
- **Workload:** each goroutine gets the lock 10,000 times. Inside the critical section, it increments a shared counter and does a small loop (100 iterations), then unlocks.
- **Waiting time:** measured from right before `Lock()` is called to right after it returns, using `time.Now()` and `time.Since()`.
- **Trials:** each setup runs 5 times, and the results are averaged.
- **Start signal:** all goroutines wait on a channel and start at the same time.
- **Same setup for both locks:** both locks use the same `bench` function, workload, and settings.
- **Correctness check:** after each trial, the counter must equal `goroutines × 10,000`. If it doesn't, the lock let two goroutines in at once.
- **Metrics:** average wait time and p99 wait time (the wait that 99% of acquisitions are faster than). p99 shows the worst-case waits.

### Results

All times are in nanoseconds, averaged over 5 trials.

| Lock | Goroutines | Avg wait (ns) | p99 wait (ns) | Correct |
|---|---|---|---|---|
| ticket | 1 | 38 | 59 | true |
| cas | 1 | 29 | 42 | true |
| ticket | 2 | 893 | 8,458 | true |
| cas | 2 | 213 | 2,275 | true |
| ticket | 4 | 17,341 | 80,084 | true |
| cas | 4 | 595 | 12,042 | true |
| ticket | 8 | 58,658 | 186,550 | true |
| cas | 8 | 2,014 | 30,358 | true |
| ticket | 16 | 125,011 | 287,483 | true |
| cas | 16 | 3,175 | 59,758 | true |
| ticket | 32 | 247,569 | 465,442 | true |
| cas | 32 | 6,199 | 99,567 | true |
## Dependencies

- Go (standard library only)
- Packages used: `fmt`, `runtime`, `sort`, `sync`, `sync/atomic`, `time`
- No external libraries. No built-in mutex is used for the locks (`sync.WaitGroup` is only used to wait for goroutines to finish).

## AI Use

We used Claude (AI chatbot) to help us understand the lock questions, check our Go lock implementations, and help structure the benchmark and README. We reviewed and tested everything ourselves and can explain all parts of the submission.
