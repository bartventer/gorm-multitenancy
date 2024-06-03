#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.3
- date: 2024-06-02

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9341460445)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1225108 | 17552 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 951297 | 16238 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 914168 | 6376 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 639678 | 5077 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1636187 | 14718 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1411664 | 13658 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1688408 | 12242 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1482774 | 11299 | 185 |
