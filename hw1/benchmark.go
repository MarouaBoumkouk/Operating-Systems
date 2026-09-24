// BENCHMARK: compares HW0 (processes) vs HW1 (goroutines).
// How it works:
//   - For each size n (how many numbers to send), run each version 10 times.
//   - Printing is OFF while timing (printing is very slow and would hide
//     the thing we want to measure).
//   - Both versions are timed the SAME way:
//         start timer -> run the whole version -> stop timer
//     So the time includes: creating the processes/goroutines,
//     sending all n numbers, and waiting for them to finish.
//   - We alternate process, goroutine, process, goroutine, ...
//     so if the computer gets busy, both versions are affected equally.
//   - We print the average, the standard deviation (how much the runs vary),
//     and the fastest run. Every single run is also saved to results.csv.

package main

import (
	"fmt"  // printing
	"math" // math.Sqrt for the standard deviation
	"os"   // creating the results.csv file
	"time" // timing
)

// The sizes we test. n = 1 shows mostly the STARTUP cost.
// n = 100000 shows mostly the cost of SENDING numbers.
var sizes = []int{1, 1000, 10000, 100000}

// How many times we repeat each version for each size.
var runs = 10

func runBenchmark() {
	// Create the CSV file and write the header line.
	file, _ := os.Create("results.csv")
	defer file.Close() // "defer" = do this when the function ends
	fmt.Fprintln(file, "version,n,run,seconds")

	// Warm-up: the very first run is slower (the program is loading),
	// so we do one run of each and don't count it.
	runProcesses(100, false)
	runGoroutines(100, false)

	fmt.Println("n        version      average(s)   stdev(s)     fastest(s)")

	// "for _, n := range sizes" = for each value n in the list sizes
	for _, n := range sizes {

		// empty lists to collect the times
		processTimes := []float64{}
		goroutineTimes := []float64{}

		for run := 1; run <= runs; run++ {

			// ---- time the process version ----
			start := time.Now()
			runProcesses(n, false)
			seconds := time.Since(start).Seconds()
			processTimes = append(processTimes, seconds) // add to the list
			fmt.Fprintf(file, "process,%d,%d,%f\n", n, run, seconds)

			// ---- time the goroutine version ----
			start = time.Now()
			runGoroutines(n, false)
			seconds = time.Since(start).Seconds()
			goroutineTimes = append(goroutineTimes, seconds)
			fmt.Fprintf(file, "goroutine,%d,%d,%f\n", n, run, seconds)
		}

		// print one line for each version
		fmt.Printf("%-8d process      %.6f     %.6f     %.6f\n",
			n, average(processTimes), stdev(processTimes), fastest(processTimes))
		fmt.Printf("%-8d goroutine    %.6f     %.6f     %.6f\n",
			n, average(goroutineTimes), stdev(goroutineTimes), fastest(goroutineTimes))
		fmt.Printf("         -> process is %.1fx slower\n\n",
			average(processTimes)/average(goroutineTimes))
	}

	fmt.Println("All runs passed the correctness check. Every run is saved in results.csv")
}

// average = add up all the times, divide by how many there are
func average(times []float64) float64 {
	total := 0.0
	for _, t := range times {
		total += t
	}
	return total / float64(len(times))
}

// stdev = standard deviation: how spread out the times are.
// Small stdev = the runs were consistent.
func stdev(times []float64) float64 {
	avg := average(times)
	total := 0.0
	for _, t := range times {
		total += (t - avg) * (t - avg)
	}
	return math.Sqrt(total / float64(len(times)-1))
}

// fastest = the smallest time in the list
func fastest(times []float64) float64 {
	best := times[0]
	for _, t := range times {
		if t < best {
			best = t
		}
	}
	return best
}
