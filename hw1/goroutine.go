// HW1 VERSION: producer and consumer are 2 GOROUTINES.
// Synchronization is the same as HW0:
//   producer sends a number -> consumer receives it -> consumer sends ack
//   -> producer receives ack -> producer sends the next number

package main

import (
	"fmt"  // printing
	"os"   // os.Exit if something is wrong
	"sync" // WaitGroup: lets main wait for goroutines to finish
)

// runGoroutines creates the channels, starts the 2 goroutines,
// and waits until both are done.
func runGoroutines(n int, print bool) {
	// make(chan int) creates a channel that carries int values.
	data := make(chan int) // producer -> consumer
	ack := make(chan int)  // consumer -> producer

	// A WaitGroup is a counter: Add(2) = "wait for 2 goroutines".
	// Each goroutine calls Done() when it finishes, and Wait() blocks
	// until the counter gets back to 0. (Like cmd.Wait() in HW0.)
	var wg sync.WaitGroup
	wg.Add(2)

	// Start both goroutines. The "go" keyword runs the function
	// in the background, so these 2 lines return right away.
	// &wg passes the WaitGroup itself (not a copy), so Done() works.
	go producer(data, ack, n, print, &wg)
	go consumer(data, ack, n, print, &wg)

	wg.Wait() // wait here until both goroutines call Done()
}

// producer sends 1, 2, ..., n and waits for an ack after each one.
func producer(data chan int, ack chan int, n int, print bool, wg *sync.WaitGroup) {
	for i := 1; i <= n; i++ {

		if print {
			fmt.Println("Producer:", i)
		}

       data <- i // send the number (waits until the consumer receives it)

		<-ack // WAIT here until the consumer sends an ack
	}
	wg.Done() // tell main "producer is finished"
}

// consumer receives n numbers and sends an ack after each one.
func consumer(data chan int, ack chan int, n int, print bool, wg *sync.WaitGroup) {
	for i := 1; i <= n; i++ {
		number := <-data // WAIT here until a number arrives

		if print {
			fmt.Println("Consumer:", number)
		}

		// Correctness check: numbers must arrive in order 1, 2, 3, ...
		if number != i {
			fmt.Println("ERROR: wrong number:", number, "expected:", i)
			os.Exit(1)
		}

		ack <- 1 // send the ack: "got it, send the next one"
	}
	wg.Done() // tell main "consumer is finished"
}
