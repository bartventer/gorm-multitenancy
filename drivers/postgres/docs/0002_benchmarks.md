#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.4
- date: 2024-06-17

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9542963354)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1241025 | 17550 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 916737 | 16231 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 929993 | 6376 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 651817 | 5076 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1676097 | 14720 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1399377 | 13653 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1690510 | 12240 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1506418 | 11297 | 185 |
