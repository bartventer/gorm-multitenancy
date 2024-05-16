#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.3
- date: 2024-05-16

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9112993515)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1178547 | 17545 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 885363 | 16069 | 208 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 949850 | 6376 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 674684 | 4917 | 86 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1654785 | 14717 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1369367 | 13496 | 204 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1664457 | 12233 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1435526 | 10977 | 183 |
