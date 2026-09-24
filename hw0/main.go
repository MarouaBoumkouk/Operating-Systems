package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "producer":
			runProducer()
			return
		case "consumer":
			runConsumer()
			return
		}
	}
	runLauncher()
}

// runLauncher creates two pipes and spawns producer/consumer as separate OS processes.
func runLauncher() {
	dataR, dataW, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	ackR, ackW, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// Producer gets: fd3=dataW (write numbers), fd4=ackR (read acks)
	prodCmd := exec.Command(exePath, "producer")
	prodCmd.ExtraFiles = []*os.File{dataW, ackR}
	prodCmd.Stdout = os.Stdout
	prodCmd.Stderr = os.Stderr

	// Consumer gets: fd3=dataR (read numbers), fd4=ackW (write acks)
	consCmd := exec.Command(exePath, "consumer")
	consCmd.ExtraFiles = []*os.File{dataR, ackW}
	consCmd.Stdout = os.Stdout
	consCmd.Stderr = os.Stderr

	if err := prodCmd.Start(); err != nil {
		panic(err)
	}
	if err := consCmd.Start(); err != nil {
		panic(err)
	}

	// Parent doesn't use the pipes itself; close its copies so fds aren't leaked.
	dataR.Close()
	dataW.Close()
	ackR.Close()
	ackW.Close()

	if err := prodCmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "producer error:", err)
	}
	if err := consCmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "consumer error:", err)
	}
}

// runProducer generates 1..5, writing each to the data pipe and waiting for
// an ack on the ack pipe before producing the next number.
func runProducer() {
	dataW := os.NewFile(3, "dataW")
	ackR := os.NewFile(4, "ackR")
	defer dataW.Close()
	defer ackR.Close()

	buf := make([]byte, 1)
	for i := 1; i <= 5; i++ {
		buf[0] = byte(i)
		if _, err := dataW.Write(buf); err != nil {
			panic(err)
		}
		fmt.Printf("Producer: %d\n", i)

		// Block until consumer acknowledges it printed this number.
		if _, err := ackR.Read(buf); err != nil {
			panic(err)
		}
	}
}

// runConsumer reads 1..5 from the data pipe, printing each, then acks so
// the producer can proceed.
func runConsumer() {
	dataR := os.NewFile(3, "dataR")
	ackW := os.NewFile(4, "ackW")
	defer dataR.Close()
	defer ackW.Close()

	buf := make([]byte, 1)
	for i := 1; i <= 5; i++ {
		if _, err := dataR.Read(buf); err != nil {
			panic(err)
		}
		fmt.Printf("Consumer: %d\n", buf[0])

		if _, err := ackW.Write(buf); err != nil {
			panic(err)
		}
	}
}
