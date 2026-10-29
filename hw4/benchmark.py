# benchmark.py  compare first, best, and worst fit
# For each workload we make ONE random list of operations (same seed),
# then run that same list on all 3 policies with the same heap size.
# Run: python3 benchmark.py

import random
from allocator import Allocator

HEAP_SIZE = 10000


def make_ops(n, min_size, max_size, malloc_chance):
    """Make a list of n operations: ("malloc", id, size) or ("free", id)."""
    random.seed(1)                           # same seed -> same workload every time
    ops = []
    alive = []                               # ids that are allocated right now
    for i in range(n):
        if not alive or random.random() < malloc_chance:
            ops.append(("malloc", i, random.randint(min_size, max_size)))
            alive.append(i)
        else:
            victim = alive.pop(random.randrange(len(alive)))   # free a random block
            ops.append(("free", victim))
    return ops


def run(ops, policy):
    a = Allocator(HEAP_SIZE, policy)
    addr_of = {}                             # id -> address
    frag = []
    for k, op in enumerate(ops):
        if op[0] == "malloc":
            addr = a.malloc(op[2])
            if addr is not None:
                addr_of[op[1]] = addr
        elif op[1] in addr_of:               # skip free if that malloc failed
            a.free(addr_of.pop(op[1]))
        if k % 100 == 0:
            frag.append(a.fragmentation())   # check fragmentation every 100 ops
    return (a.searched / a.mallocs,          # avg blocks searched per malloc
            100 * a.failed / a.mallocs,      # % of mallocs that failed
            100 * sum(frag) / len(frag),     # avg fragmentation %
            len(a.free_list))                # number of free blocks at the end


# name, min size, max size, chance of malloc, number of ops
workloads = [
    ("Small (1-20), 50% malloc",     1,  20, 0.5, 10000),
    ("Large (100-800), 50% malloc", 100, 800, 0.5, 10000),
    ("Mixed (1-500), 50% malloc",    1, 500, 0.5, 10000),
    ("Small (1-20), 70% malloc",     1,  20, 0.7, 10000),
]

print("| Workload | Policy | Avg blocks searched | Failed % | Avg fragmentation % | Free blocks at end |")
print("|---|---|---|---|---|---|")
for name, lo, hi, p, n in workloads:
    ops = make_ops(n, lo, hi, p)
    for policy in ["first", "best", "worst"]:
        s, f, fr, nb = run(ops, policy)
        print("| %s | %s | %.1f | %.1f | %.1f | %d |" % (name, policy, s, f, fr, nb))