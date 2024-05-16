#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: Intel(R) Core(TM) i7-10750H CPU @ 2.60GHz
- go version: 1.22.3
- date: 2024-05-16



The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 43935165 | 17795 | 225 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 48148485 | 16101 | 208 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 485293 | 6374 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 159456 | 4917 | 86 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 44431652 | 14847 | 210 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 47528565 | 13601 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 45509287 | 12320 | 187 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 47675177 | 11057 | 180 |
