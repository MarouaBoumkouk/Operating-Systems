# EECE 4811/5811 HW1

## Group Members
* Maroua Boumkouk
* Oluchukwu Okeke

## Q1 – Threads and Lightweight Concurrency

### OS Threads

Threads are smaller units of work that run inside a process. Multiple threads in the same process share the same code, data, and memory, but each thread has its own stack, registers, and program counter. The program counter keeps track of which instruction that thread is currently executing [1]. Threads are useful because they can make programs more responsive, allow multiple tasks to happen concurrently, and make better use of multiple CPU cores [1].

There are user-level threads and kernel-level threads. User-level threads are managed by a thread library instead of directly by the operating system. They are lightweight, but one limitation is that if a user-level thread makes a blocking system call, the entire process may become blocked. Kernel-level threads are managed and scheduled directly by the operating system. If one kernel thread becomes blocked, the OS can schedule another thread from the same process. However, kernel-level threads require more interaction with the operating system, which can add more overhead [1].

One major difference between a process and a thread is memory. Separate processes normally have separate memory spaces, while threads inside the same process share memory and other resources [1].

### Go Goroutines

A goroutine is a lightweight unit of execution managed by the Go runtime instead of directly by the operating system [2]. Goroutines allow functions to run concurrently and usually require less memory and have a faster startup time than traditional OS threads [2].

Every Go program begins with a main goroutine. A normal function call such as `display()` runs inside the current goroutine and must finish before the program continues past that call. If we write `go display()`, Go starts that function as another goroutine, allowing the main goroutine and the new goroutine to make progress concurrently [2].

One limitation is that the order in which different goroutines run is not always predictable. Because several goroutines may access shared information at the same time, synchronization is often necessary. This can also make debugging more difficult [2].

The example source uses `time.Sleep()` to give another goroutine enough time to run before the main goroutine finishes. However, `Sleep()` is mainly useful for demonstrating the idea. In a real program, proper synchronization should be used when we need to guarantee that one goroutine waits for another.

### Java Virtual Threads

Java has platform threads and virtual threads. A platform thread is closely connected to an operating-system thread. A platform thread acts as a wrapper around an OS thread and normally uses that OS thread for its lifetime [3][4].

A virtual thread is a lightweight Java thread managed by the Java runtime instead of being permanently tied to one specific OS thread. The Java runtime can schedule many virtual threads onto a smaller number of platform or OS threads [4].

This means that if one virtual thread is waiting for something such as a network request or database operation, the underlying OS thread may be used to perform work for another virtual thread. This allows Java programs to support a very large number of concurrent tasks without needing the same number of OS threads [4].

Virtual threads are especially useful for programs that spend a lot of time waiting for I/O, such as server applications, HTTP requests, and database queries [3][4]. However, virtual threads do not make CPU-intensive work execute faster. Their main goal is to improve scalability and throughput when many concurrent tasks are running [3][4].

### Comparison and Tradeoffs

The three approaches differ mainly in who creates and schedules them.

**OS/kernel threads** are managed and scheduled directly by the operating system.

**Go goroutines** are created and managed by the Go runtime. The Go runtime uses OS threads underneath to actually execute the goroutines.

**Java virtual threads** are created and scheduled by the Java runtime. The Java scheduler assigns virtual threads to platform threads, and the operating system then schedules those platform threads onto the CPU [4].

Lightweight concurrency mechanisms such as goroutines and Java virtual threads have several advantages. They use fewer resources than creating large numbers of traditional OS threads, so programs can have many more concurrent tasks. This is especially useful for applications that spend a lot of time waiting for operations such as network requests, files, or databases.

However, lightweight concurrency also introduces some limitations. The order in which concurrent tasks execute may be unpredictable. Programs may need synchronization when different tasks access shared data. Debugging concurrent programs can also be more difficult.

I would prefer traditional OS or platform threads when I have a smaller number of threads or work that needs direct operating-system scheduling and significant CPU execution. I would prefer goroutines when writing a Go program that needs many concurrent tasks. I would prefer Java virtual threads when writing a Java application that needs to handle many tasks that spend a large amount of time waiting, such as server requests, network operations, or database queries.

## Sources

**[1] GeeksforGeeks, “Thread in Operating System.”**
I consider this source useful because GeeksforGeeks is a well-known educational computer science website. The article explains thread structure, user-level threads, kernel-level threads, and the differences between processes and threads.

**[2] GeeksforGeeks, “Goroutines – Concurrency in Golang.”**
I consider this source useful because it provides explanations and code examples showing how goroutines are created with the `go` keyword, how the main goroutine behaves, and some advantages and limitations of goroutines.

**[3] GeeksforGeeks, “Virtual Threads in Java.”**
I consider this source useful because it explains the differences between Java platform threads and virtual threads and provides examples showing how virtual threads are created and used.

**[4] OpenJDK, JEP 444: “Virtual Threads,” Ron Pressler and Alan Bateman.**
I consider this a primary and credible source because it is an official OpenJDK design document written by developers involved with Java virtual threads. It explains why virtual threads were created, how they are scheduled, their relationship with OS threads, and the design goals and limitations of virtual threads.

## AI Assistance

We used OpenAI ChatGPT to help us understand OS threads, Go goroutines, and Java virtual threads in simpler language. We also used ChatGPT to check our technical understanding, correct grammar, organize our explanation.


## Q2 – Code (Process vs. Goroutine)

### Files

| File | What it does |
|---|---|
| `main.go` | Reads the command (`process`, `goroutine`, `bench`) and runs that part |
| `process.go` | **HW0**: producer and consumer as 2 separate **processes**, talking through 2 **pipes** |
| `goroutine.go` | **HW1**: producer and consumer as 2 **goroutines**, talking through 2 **channels** |
| `benchmark.go` | Times both versions many times and compares them |
| `results.csv` | Raw timing data from our benchmark run |

### Dependencies
- Go 1.20 or newer (https://go.dev/dl/)
- Only the Go standard library, no extra packages

### How to run
Run these inside the `hw1` folder:

```bash
go run . process      # HW0 demo: 5 numbers with processes (prints each one)
go run . goroutine    # HW1 demo: 5 numbers with goroutines (prints each one)
go run . bench        # benchmark: compares both, printing off (~20 s)
```

### Program design

**Both versions use the same protocol (from HW0):**
- A **data** line (producer → consumer) carries the number.
- An **ack** line (consumer → producer) says "got it."
- The producer sends a number, then **waits for the ack** before sending the next one. So producer and consumer go in **lockstep**, one number at a time.
- The consumer checks that the numbers arrive in order (1, 2, 3, …). If not, the program stops with an error.

**HW0 – processes (`process.go`)**
- The parent makes 2 pipes, then starts this same program 2 more times: once as `producer`, once as `consumer`. It gives them the pipe ends as file descriptors 3 and 4.
- Reading from an empty pipe **blocks**, and that's how each side waits for the other.
- The parent waits for both child processes to finish (`Wait()`).

**HW1 – goroutines (`goroutine.go`)**
- Everything runs in 1 process. `go producer(...)` and `go consumer(...)` start the 2 goroutines.
- The 2 pipes are replaced by 2 **unbuffered channels**: `data` and `ack`. A send (`data <- i`) waits until the other side receives (`<-data`). This gives the same blocking behavior as the pipes.
- A `sync.WaitGroup` waits for both goroutines to finish, just like `Wait()` in HW0.

### How we measured
- We timed each version with `time.Now()` in `benchmark.go`: start timer → run → stop timer.
- Sizes: N = 1, 1,000, 10,000, 100,000 numbers. N = 1 shows startup cost, N = 100,000 shows message cost.
- 10 runs each, printing off, 1 warm-up run first. All runs are saved in `results.csv`.

### What we kept the same
Same language, same laptop, same numbers, same send → wait-for-ack protocol.

### Results (MacBook Pro, average of 10 runs)

| N | Process (s) | Goroutine (s) | Process is … slower |
|---:|---:|---:|---:|
| 1 | 0.00230 | 0.000017 | 138x |
| 1,000 | 0.00741 | 0.000229 | 32x |
| 10,000 | 0.0596 | 0.00215 | 28x |
| 100,000 | 0.560 | 0.0209 | 27x |

### Why goroutines were faster
Based on our test, goroutines are faster than processes. When we ran both versions and compared the results, the process version was about 27x slower with 100,000 numbers, and about 138x slower with just 1 number.

For the messages, the goroutines send numbers through a channel that we created inside our program, but the processes send them through a pipe. The pipe goes through the operating system every time, which is slower. The channel stays inside our program, so it's faster.

For the startup, the parent process has to ask the operating system to create two new child processes (producer and consumer), and each one is a whole new program with its own memory. A goroutine is much smaller and runs inside the same program, so it starts much faster.
## AI Assistance
We used Claude (AI) for help with Go syntax and debugging while working on Q2. 