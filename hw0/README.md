# HW0 — Producer/Consumer IPC via OS Pipes

## Q1: Reading Questions

1. Two processes cannot directly access each other's variables because each runs in its own isolated virtual address space.
 An address that is valid in one process has no meaningful connection to the same address in another.
 The operating system enforces this separation for security and stability. To exchange data,
 processes must use an explicit interprocess communication method, such as pipes, shared memory, or sockets.

2. Data flows one way only: writes to the write end come out the read end.
 It's a unidirectional byte stream — that's why two-way communication needs two pipes.

3. Reading from an empty pipe blocks — the process just pauses there until data shows up.
 That's handy for synchronization: instead of polling or sleeping in a loop,
 the process can block on the read and the OS will wake it up the moment data's actually available.

4.Closing unused ends matters mainly for EOF: if a write end stays open somewhere,
 even unused, the pipe never signals EOF, so the reader can end up blocked forever waiting for data that's not coming.
 Closing unused ends also frees up file descriptors and keeps it clear which process actually owns which end.
 
5.No — the pipe only guarantees the numbers arrive at the Consumer in order, not when either side prints.
 Without extra synchronization, Producer could dump 1–5 into the pipe (and print them) before Consumer even runs.
 That's why this design needs a second, ack pipe: it makes Producer block until Consumer confirms each number was processed



## How to Compile/Run

Requires Go (tested on go1.21+).

    go run main.go

Or build a binary:

    go build -o hw0
    ./hw0

## Design

The program has three roles selected via a subcommand argument, all handled
in a single binary (main.go):

- **Launcher** (no args): creates two OS pipes and spawns Producer and
  Consumer as separate child processes using `exec.Command`, passing pipe
  file descriptors via `ExtraFiles`.
- **Producer**: writes numbers 1–5 to the data pipe, printing each as it's
  sent. After each write it blocks reading the ack pipe, so it cannot send
  the next number until Consumer confirms receipt of the current one.
- **Consumer**: reads each number from the data pipe, prints it, then
  writes a single byte back on the ack pipe to unblock Producer.

Two pipes are required because a pipe is unidirectional: one pipe carries
the data (Producer → Consumer), the other carries synchronization acks
(Consumer → Producer). The ack pipe is what enforces the strict
Producer/Consumer print alternation without using sleep().

Both processes close their pipe file descriptors and exit after the 5th
number/ack, and the launcher waits on both child processes before exiting.

## Dependencies

None beyond the Go standard library (os, os/exec, fmt).
