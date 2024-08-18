## Hash Function for Lock Key Generation

### Objective

The objective of this document is to explain the rationale behind using the FNV-1a hash function for generating lock keys in the Gorm Multitenancy library. We will also provide benchmark results comparing the performance of FNV-1a to other hashing functions.

### Rationale

We opted for the FNV-1a hash function for generating lock keys due to the following reasons:
- **Speed**: FNV-1a is faster compared to other hashing functions.
- **Simplicity**: It is simpler to implement and use.
- **Efficiency**: Directly produces a 64-bit integer, suitable for PostgreSQL and MySQL advisory locks.
- **Low Collision Rate**: Designed to maintain a low collision rate, making it well-suited for hash tables, checksums, and data fingerprinting.

### Benchmark Results

We conducted benchmarks to compare the performance of various hashing functions for generating lock keys. The metrics used are:

- **Time (ns/op)**: Time taken per operation in nanoseconds.
- **Memory (B/op)**: Memory allocated per operation in bytes.
- **Allocations (allocs/op)**: Number of memory allocations per operation.

#### Results

| Function | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
| --- | --- | --- | --- |
| FNV-1a ([`GenerateLockKey`](../pkg/migrator/locking.go)) | 5.927 | 0 | 0 |
| MD5 ([`GenerateLockKeyMD5`](../pkg/migrator/locking_test.go)) | 168.9 | 48 | 1 |
| SHA-1 ([`GenerateLockKeySHA1`](../pkg/migrator/locking_test.go)) | 206.7 | 64 | 1 |
| SHA-256 ([`GenerateLockKeySHA256`](../pkg/migrator/locking_test.go)) | 269.9 | 64 | 1 |
| SHA-512 ([`GenerateLockKeySHA512`](../pkg/migrator/locking_test.go)) | 349.5 | 96 | 1 |

Based on these results, FNV-1a is significantly more efficient in terms of both time and memory usage.

### Running the Benchmarks

To run the benchmarks on your own machine, follow these steps:

1. **Clone the repository**:
    ```sh
    git clone https://github.com/bartventer/gorm-multitenancy.git
    cd gorm-multitenancy
    ```

2. **Run the benchmarks**:
    ```sh
    go test -benchmem -run=^$ -bench ^BenchmarkGenerateLockKey$ github.com/bartventer/gorm-multitenancy/v8/pkg/migrator -v
    ```

