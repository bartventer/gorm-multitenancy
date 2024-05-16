#### Benchmarks

The benchmarks were run with the following configuration:

- goos: {{GOOS}}
- goarch: {{GOARCH}}
- pkg: {{PKG}}
- cpu: {{CPU}}
- go version: {{GO_VERSION}}
- date: {{DATE}}

{{RUNNER_TEXT}}

The following table shows the benchmark results, obtained by running:
```bash
{{CMD}}
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
{{BENCHMARKS}}