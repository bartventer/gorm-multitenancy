//go:build gorm_multitenancy_benchmarks
// +build gorm_multitenancy_benchmarks

package drivertest

import "testing"

func BenchmarkConformance(b *testing.B) {
	RunConformanceBenchmarks(b, newHarness)
}
