#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.3
- date: 2024-06-03

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9350984997)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1153284 | 17557 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 871816 | 16227 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 944979 | 6377 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 667537 | 5076 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1601039 | 14720 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1315298 | 13656 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1615386 | 12235 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1419866 | 11300 | 185 |
