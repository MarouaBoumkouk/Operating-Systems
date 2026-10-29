# EECE 4811/5811 HW4

## Group Members
* Maroua Boumkouk
* Oluchukwu Okeke

## CPU Scheduling (3 pt)
**Q1. Three jobs of length 200, SJF and FIFO**
all three jobs arrive at time 0 and each runs for 200 seconds. Under FIFO, the jobs finish at 200, 400, and 600, so the average turnaround time is 400. They start at 0, 200, and 400, so the average response time is 200. SJF gives the same answer because all jobs have equal lengths, so sorting them doesn’t change the order.

| Job | Response | Turnaround |
|---|---|---|
| 0 | 0 | 200 |
| 1 | 200 | 400 |
| 2 | 400 | 600 |
| **Avg** | **200** | **400** |

**Q2. Jobs of length 100, 200, 300**
Command: `python3 scheduler.py -l 100,200,300 -p FIFO -c` (same with `-p SJF`)
| Job | Length | Response | Turnaround |
|---|---|---|---|
| 0 | 100 | 0 | 100 |
| 1 | 200 | 100 | 300 |
| 2 | 300 | 300 | 600 |
| **Avg** | | **133.33** | **333.33** |
FIFO and SJF give the same result because the jobs already arrive
shortest first, so SJF doesn't change the order.

**Q3. Same, plus RR with time slice 1**
| Job | Length | Response | Turnaround |
|---|---|---|---|
| 0 | 100 | 0 | 298 |
| 1 | 200 | 1 | 499 |
| 2 | 300 | 2 | 600 |
| **Avg** | | **1.00** | **465.67** |

RR gives every job a fast start, so response is only 1. But jobs share
the CPU, so they finish later, and turnaround (465.67) is worse than
SJF (333.33).

**Q4. When does SJF give the same turnaround as FIFO?**
When the jobs arrive already sorted from shortest to longest (or all jobs are the same length). Then SJF doesn't reorder anything, so it runs the exact same schedule as FIFO.

**Q5. Types of workloads and quantum lengths**
Assuming all jobs arrive at time 0, SJF and RR have the same response times when jobs are ordered shortest to longest and the quantum lets every job before the last finish in one turn. A quantum at least as large as the longest job is sufficient. In that case, RR never cuts a job off, so it behaves like FIFO, and FIFO is the same as SJF when jobs are sorted.

## Virtual and Physical Addresses (3 pt)

### Design (`allocator.py`)
 
- The heap is fake: just addresses 0 to 9999 (heap size 10000). No real memory is used.
- `free_list` is a list of free blocks, each `[address, size]`. `used` is a dictionary `{address: size}` for allocated blocks.
- **malloc(size):** looks through the free list and picks a block using the policy:
  - **First fit:** takes the first block that is big enough, then stops looking.
  - **Best fit:** checks every block and takes the smallest one that fits.
  - **Worst fit:** checks every block and takes the biggest one.
- **Splitting:** if the block is bigger than needed, the user gets the front part and the rest stays free.
  Example: free block `[0, 100]`, `malloc(10)` returns 0 and the free list becomes `[10, 90]`.
- **free(addr):** puts the block back, sorts the list by address, then **coalesces**: if one block ends exactly where the next one starts, they merge into one.
  Example: `[0, 10]` and `[10, 20]` merge into `[0, 30]`.
- `python3 allocator.py` runs a small demo that shows splitting and coalescing.
### Benchmark methodology (`benchmark.py`)
 
- Each workload is a random list of 10,000 mallocs and frees. Frees pick a random allocated block, which creates holes.
- The random seed is fixed, so **all three policies get the exact same list of operations and the same heap size**. That keeps the comparison fair.
- If a malloc fails, its free is skipped.
**What we measure**
 
| Metric | Meaning |
|---|---|
| Avg blocks searched | Performance overhead: how many free blocks malloc looks at each time |
| Failed % | Memory efficiency: % of mallocs with no block big enough |
| Avg fragmentation % | `1 - largest free block / total free space`, checked every 100 ops. 0% = one big free block, high = lots of small pieces |
| Free blocks at end | How many holes are left in the heap |
 
**Workloads and why we picked them**
 
| Workload | Why |
|---|---|
| Small (1-20), 50% malloc | Heap never fills up, so it shows search cost and how each policy leaves holes |
| Large (100-800), 50% malloc | Big requests put pressure on the heap and show failures |
| Mixed (1-500), 50% malloc | Many different sizes, so lots of different hole sizes |
| Small (1-20), 70% malloc | More mallocs than frees, so the heap fills up and the free list gets long |
 
### Results
 
| Workload | Policy | Avg blocks searched | Failed % | Avg fragmentation % | Free blocks at end |
|---|---|---|---|---|---|
| Small (1-20), 50% malloc | first | 14.7 | 0.0 | 2.5 | 63 |
| Small (1-20), 50% malloc | best | 22.9 | 0.0 | 2.3 | 35 |
| Small (1-20), 50% malloc | worst | 42.7 | 0.0 | 68.2 | 93 |
| Large (100-800), 50% malloc | first | 13.1 | 62.2 | 77.1 | 19 |
| Large (100-800), 50% malloc | best | 15.5 | 62.4 | 71.7 | 16 |
| Large (100-800), 50% malloc | worst | 15.5 | 63.9 | 82.5 | 22 |
| Mixed (1-500), 50% malloc | first | 24.1 | 49.8 | 75.5 | 35 |
| Mixed (1-500), 50% malloc | best | 27.9 | 49.0 | 72.9 | 37 |
| Mixed (1-500), 50% malloc | worst | 29.8 | 51.9 | 85.0 | 33 |
| Small (1-20), 70% malloc | first | 96.9 | 49.6 | 76.3 | 139 |
| Small (1-20), 70% malloc | best | 51.6 | 49.7 | 71.7 | 47 |
| Small (1-20), 70% malloc | worst | 342.1 | 50.5 | 90.1 | 498 |
 
### Analysis
 
- **First fit** does the least searching when the heap isn't full, because it stops at the first block that fits (14.7 vs 22.9 for best fit on small requests). But it leaves small leftover pieces at the front of the list. When the heap fills up (70% malloc), it has to walk past all of them, and its search cost goes up to 96.9.
- **Best fit** leaves the fewest holes, because it picks the tightest fit, so the leftover piece is small. It has the lowest fragmentation in almost every test. Its free list stays short, so in the 70% malloc test it actually searches less than first fit (51.6 vs 96.9). The downside is that it checks the whole list every time.
- **Worst fit** is the worst here. It always cuts from the biggest block, so big blocks get broken up quickly. In the small test its fragmentation is 68% vs about 2% for the others. In the 70% test it ends with 498 holes and searches 342 blocks per malloc.
- **With big requests** (Large and Mixed), all three fail about the same amount (49-64%) because the heap is just too small. Best fit is still slightly better on fragmentation.
**Tradeoff:** first fit is the fastest in normal use, best fit uses memory the best and keeps the list short, and worst fit is bad at both in our tests.
 
---
 
## Sources
 
- OSTEP Ch. 7, Scheduling: https://pages.cs.wisc.edu/~remzi/OSTEP/cpu-sched.pdf
- OSTEP Ch. 17, Free-Space Management: https://pages.cs.wisc.edu/~remzi/OSTEP/vm-freespace.pdf
- Simulator: https://github.com/remzi-arpacidusseau/ostep-homework/tree/master/cpu-sched
## AI use
We used Claude (Anthropic); We checked the scheduler answers by running the simulator and reran all the benchmarks ourselves.

