# allocator.py  list memory allocator
# We pretend the heap is addresses 0 to heap_size-1.
# free_list = list of free blocks, each one is [address, size]
# used      = dictionary {address: size} of allocated blocks


class Allocator:
    def __init__(self, heap_size, policy):
        self.policy = policy                 # "first", "best", or "worst"
        self.free_list = [[0, heap_size]]    # at the start, the whole heap is free
        self.used = {}
        self.searched = 0                    # how many blocks we looked at (overhead)
        self.mallocs = 0
        self.failed = 0

    def find_block(self, size):
        """Pick a free block using the policy. Returns its index or None."""
        fits = []                            # indexes of blocks that are big enough
        for i in range(len(self.free_list)):
            self.searched += 1
            if self.free_list[i][1] >= size:
                if self.policy == "first":
                    return i                 # first fit: take the first one, stop
                fits.append(i)
        if not fits:
            return None
        if self.policy == "best":            # best fit: smallest block that fits
            return min(fits, key=lambda i: self.free_list[i][1])
        else:                                # worst fit: biggest block
            return max(fits, key=lambda i: self.free_list[i][1])

    def malloc(self, size):
        self.mallocs += 1
        i = self.find_block(size)
        if i is None:
            self.failed += 1                 # no block big enough
            return None
        addr, block_size = self.free_list[i]
        # SPLITTING: give the first `size` units to the user,
        # the rest stays in the free list
        if block_size == size:
            self.free_list.pop(i)
        else:
            self.free_list[i] = [addr + size, block_size - size]
        self.used[addr] = size
        return addr

    def free(self, addr):
        size = self.used.pop(addr)
        self.free_list.append([addr, size])
        self.free_list.sort()                # keep blocks in address order
        # COALESCING: if a block ends where the next one starts, merge them
        merged = []
        for a, s in self.free_list:
            if merged and merged[-1][0] + merged[-1][1] == a:
                merged[-1][1] += s
            else:
                merged.append([a, s])
        self.free_list = merged

    def fragmentation(self):
        """External fragmentation = 1 - (largest free block / total free).
        0 means all free space is in one piece. Near 1 means lots of tiny pieces."""
        total = sum(s for a, s in self.free_list)
        if total == 0:
            return 0
        largest = max(s for a, s in self.free_list)
        return 1 - largest / total


# small demo: python3 allocator.py
if __name__ == "__main__":
    a = Allocator(100, "first")
    x = a.malloc(10)
    y = a.malloc(20)
    z = a.malloc(30)
    print("after 3 mallocs:", a.free_list)   # one block left: [60, 40]
    a.free(y)
    print("free y:", a.free_list)            # hole at 10
    a.free(x)
    print("free x:", a.free_list)            # x and y merge into [0, 30]
    a.free(z)
    print("free z:", a.free_list)            # everything merges back to [0, 100]