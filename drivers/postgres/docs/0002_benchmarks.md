#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.3
- date: 2024-06-02

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9341309653)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1164300 | 17550 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 890558 | 16231 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 928983 | 6376 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 656954 | 5077 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1619777 | 14721 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1383346 | 13657 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1656402 | 12232 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1421472 | 11307 | 185 |
