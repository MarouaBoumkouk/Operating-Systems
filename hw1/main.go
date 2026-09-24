// EECE 4811/5811 HW1
// How to run:
//   go run . process     HW0 demo (sends 5 numbers and prints them)
//   go run . goroutine   HW1 demo (sends 5 numbers and prints them)
//   go run . bench       runs the benchmark

package main

// "import" = the libraries we use (all built into Go)
import (
	"fmt"     // printing
	"os"      // os.Args = the words typed after the program name
	"strconv" // converts text like "100" into the number 100
)

// main() is where every Go program starts.
func main() {
	// os.Args is a list of the words on the command line.
	// Example: "go run . bench"  ->  os.Args[1] is "bench"
	// If the user typed nothing after the program name, show help and stop.
	if len(os.Args) < 2 {
		fmt.Println("usage: go run . process | goroutine | bench")
		return
	}

	command := os.Args[1]

	// "switch" is like a chain of if / else if.
	switch command {

	case "process":
		runProcesses(5, true) // HW0 demo: 5 numbers, printing ON

	case "goroutine":
		runGoroutines(5, true) // HW1 demo: 5 numbers, printing ON

	case "bench":
		runBenchmark()

	// The next 2 cases are NOT typed by the user.
	// The process version (process.go) starts this same program again
	// as a child process, like:   <program> producer 100 false
	//   os.Args[2] = how many numbers
	//   os.Args[3] = "true" or "false" (print or not)
	case "producer":
		n, _ := strconv.Atoi(os.Args[2]) // "100" -> 100 (the _ ignores the error)
		print := os.Args[3] == "true"
		runProducer(n, print)

	case "consumer":
		n, _ := strconv.Atoi(os.Args[2])
		print := os.Args[3] == "true"
		runConsumer(n, print)

	default:
		fmt.Println("unknown command:", command)
		fmt.Println("usage: go run . process | goroutine | bench")
	}
}
