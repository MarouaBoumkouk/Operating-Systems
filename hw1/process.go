// HW0 VERSION: producer and consumer are 2 separate PROCESSES.
// Synchronization (same as HW0):
//   1. producer writes a number into the data pipe
//   2. consumer reads it (reading an empty pipe WAITS until data arrives)
//   3. consumer writes an "ack" into the ack pipe
//   4. producer waits for that ack, and only then sends the next number
// So they take turns: one number at a time.

package main

import (
	"bufio"   // buffered reading (lets fmt.Fscan read from a pipe)
	"fmt"     // printing, and writing/reading numbers as text
	"os"      // pipes and files
	"os/exec" // starting new processes
	"strconv" // number -> text
)

// runProcesses is the PARENT. It:
//  1. makes the 2 pipes
//  2. starts 2 child processes (producer and consumer)
//  3. waits for both to finish
func runProcesses(n int, print bool) {

	// ---- Step 1: make 2 pipes ----
	// os.Pipe() gives back 2 ends: a reading end and a writing end.
	dataRead, dataWrite, _ := os.Pipe()
	ackRead, ackWrite, _ := os.Pipe()

	// ---- Step 2: start the 2 child processes ----
	// We start THIS SAME program again, but with the word "producer"
	// or "consumer", so main.go knows which job to do.
	myProgram, _ := os.Executable() // path to this program
	nText := strconv.Itoa(n)        // 100 -> "100"
	printText := strconv.FormatBool(print)

	producer := exec.Command(myProgram, "producer", nText, printText)
	consumer := exec.Command(myProgram, "consumer", nText, printText)

	// Give each child the pipe ends it needs.
	// ExtraFiles become file number 3 and 4 inside the child.
	//   producer: 3 = dataWrite (send numbers), 4 = ackRead  (get acks)
	//   consumer: 3 = dataRead  (get numbers),  4 = ackWrite (send acks)
	producer.ExtraFiles = []*os.File{dataWrite, ackRead}
	consumer.ExtraFiles = []*os.File{dataRead, ackWrite}

	// Let the children print to our screen.
	producer.Stdout = os.Stdout
	consumer.Stdout = os.Stdout
	consumer.Stderr = os.Stderr

	producer.Start() // start runs the child and does NOT wait for it
	consumer.Start()

	// The parent doesn't use the pipes itself, so it closes its copies.
	dataRead.Close()
	dataWrite.Close()
	ackRead.Close()
	ackWrite.Close()

	// ---- Step 3: wait for both children to finish ----
	producer.Wait()
	err := consumer.Wait()
	if err != nil { // the consumer exits with an error if a number was wrong
		fmt.Println("ERROR: consumer got a wrong number")
		os.Exit(1)
	}
}

// runProducer runs in the PRODUCER child process.
// It sends 1, 2, ..., n and waits for an ack after each one.
func runProducer(n int, print bool) {
	// Pick up the pipe ends the parent gave us (file 3 and 4).
	dataWrite := os.NewFile(3, "dataWrite")
	ackRead := bufio.NewReader(os.NewFile(4, "ackRead"))

	for i := 1; i <= n; i++ {
		fmt.Fprintln(dataWrite, i) // write the number into the data pipe

		if print {
			fmt.Println("Producer:", i)
		}

		var ack int
		fmt.Fscan(ackRead, &ack) // WAIT here until the consumer sends an ack
	}
	dataWrite.Close()
}

// runConsumer runs in the CONSUMER child process.
// It reads n numbers and sends an ack after each one.
func runConsumer(n int, print bool) {
	dataRead := bufio.NewReader(os.NewFile(3, "dataRead"))
	ackWrite := os.NewFile(4, "ackWrite")

	for i := 1; i <= n; i++ {
		var number int
		fmt.Fscan(dataRead, &number) // WAIT here until a number arrives

		if print {
			fmt.Println("Consumer:", number)
		}

		// Correctness check: numbers must arrive in order 1, 2, 3, ...
		if number != i {
			fmt.Fprintln(os.Stderr, "wrong number:", number, "expected:", i)
			os.Exit(1) // exit with an error so the parent knows
		}

		fmt.Fprintln(ackWrite, 1) // send the ack: "got it, send the next one"
	}
	ackWrite.Close()
}
