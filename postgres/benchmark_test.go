//go:build gorm_multitenancy_benchmarks
// +build gorm_multitenancy_benchmarks

package postgres

import (
	"testing"

	"github.com/bartventer/gorm-multitenancy/v7/pkg/drivertest"
)

func BenchmarkConformance(b *testing.B) {
	drivertest.RunConformanceBenchmarks(b, newHarness)
}
