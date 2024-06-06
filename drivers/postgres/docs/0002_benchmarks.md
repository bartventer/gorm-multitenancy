#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.3
- date: 2024-06-06

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9399013373)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1184112 | 17552 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 918555 | 16224 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 951405 | 6377 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 667583 | 5076 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1665855 | 14720 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1387175 | 13657 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1701148 | 12234 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1503335 | 11303 | 185 |
