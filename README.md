# Swiss Map 

This repository provides a highly efficient hash map implementation in Go, inspired by Swiss tables as presented by **Matt Kulukundis** from Google at [this talk](https://www.youtube.com/watch?v=ncHmEUmJZf4&t=0s) and described in Abseil's [article](https://abseil.io/blog/20180927-swisstables). The original ideas behind the Swiss map have been adapted for Go, with significant performance optimizations. This repository also takes inspiration from the work done in [Dolthub's Swiss Map](https://github.com/dolthub/swiss/) and [CockroachDB's Swiss Map](https://github.com/cockroachdb/swiss/).

## Features

- **Memory efficiency**: Uses half the memory compared to Go's native map.
- **Performance**: Significantly faster lookups and insertions due to improved hashing and control byte management.
- **Optimized for large composite keys**: Leveraging a specialized non-cryptographic hash function for better speed with complex keys.
- **Non-thread-safe**: The implementation is optimized for single-threaded use and requires external synchronization for concurrent use.

## Key Concepts

### Dual Hashing Approach

This map uses two different hashing strategies:

1. **Built-in Go map hash**: For general key types.
2. **`memhash` from the Go runtime**: This hash function is much faster for composite keys (e.g., structs), though it is not cryptographically secure.

The choice of hashing function can provide significant performance improvements for non-primitive key types, especially when dealing with large or complex keys. However, it's important to note that `memhash` is not suitable when cryptographic security is required.

### Memory and Speed Optimization

Compared to the standard Go map, this implementation is **twice as memory efficient** and **significantly faster** in terms of both insertions and lookups. The control bytes allow efficient management of empty and deleted slots, ensuring that the map can scale well with minimal memory overhead.

## Code Example

Here is how you can use the Swiss map:

```go
package main

import (
    "fmt"
    "github.com/crn4/swiss"
)

func main() {
    // Create a new map with an initial capacity
    m := swiss.New(2) 

    // Insert values
    m.Put(1, "one")
    m.Put(2, "two")

    // Retrieve values
    val, found := m.Get(1)
    if found {
        fmt.Println("Found:", val) // Output: Found: one
    }

    // Delete a value
    m.Delete(2)
    _, found = m.Get(2)
    fmt.Println("Found after delete:", found) // Output: Found after delete: false
}
```

### Benchmark results 
This map is significantly faster on large sizes and more memory-efficient than the built-in Go map. Below, you find a benchmark test to compare it with the Go native map.
```
goos: darwin
goarch: arm64
pkg: github.com/crn4/swiss
cpu: Apple M2 Pro
BenchmarkGetIntInt/runtime_map,_size:_128-12         	244543017	         4.813 ns/op	         5.000 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/swiss,_size:_128-12               	226108815	         6.292 ns/op	         2.000 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/runtime_map,_size:_1024-12        	175367342	         6.837 ns/op	        40.00 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/swiss,_size:_1024-12              	222044619	         5.427 ns/op	        20.00 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/runtime_map,_size:_16384-12       	70360771	        16.62 ns/op	       616.0 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/swiss,_size:_16384-12             	135133118	         8.723 ns/op	       312.0 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/runtime_map,_size:_131072-12      	69001207	        17.61 ns/op	      4455 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/swiss,_size:_131072-12            	100000000	        11.05 ns/op	      2473 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/runtime_map,_size:_1048576-12     	36553603	        32.77 ns/op	     39168 memalloc/kb	       0 B/op	       0 allocs/op
BenchmarkGetIntInt/swiss,_size:_1048576-12           	62780661	        17.48 ns/op	     19894 memalloc/kb	       0 B/op	       0 allocs/op

BenchmarkGetStructStruct/runtime_map,_size:_128-12         	37005960	        31.61 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/swiss,_size:_128-12               	120251827	        10.05 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/runtime_map,_size:_1024-12        	33766830	        34.92 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/swiss,_size:_1024-12              	124144640	         9.627 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/runtime_map,_size:_16384-12       	30842612	        38.07 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/swiss,_size:_16384-12             	88672134	        13.44 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/runtime_map,_size:_131072-12      	27969528	        40.97 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/swiss,_size:_131072-12            	77084702	        15.84 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/runtime_map,_size:_1048576-12     	15166243	        80.72 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetStructStruct/swiss,_size:_1048576-12           	21424888	        54.48 ns/op	       0 B/op	       0 allocs/op
```
