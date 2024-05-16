#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.3
- date: 2024-05-16

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9113233960)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1132556 | 17560 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 866528 | 16074 | 208 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 951361 | 6377 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 679585 | 4916 | 86 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1639658 | 14717 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1328368 | 13495 | 204 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1645932 | 12233 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1428971 | 10979 | 183 |
