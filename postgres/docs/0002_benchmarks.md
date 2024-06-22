#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/postgres/v7/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.4
- date: 2024-06-18

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9565315909)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/postgres/v7/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1158551 | 17548 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 892393 | 16236 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 938606 | 6375 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 667922 | 5076 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1649238 | 14719 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1362411 | 13656 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1655629 | 12240 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1438684 | 11299 | 185 |
