//go:build gorm_multitenancy_benchmarks
// +build gorm_multitenancy_benchmarks

package mysql

import (
	"testing"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/drivertest"
)

func BenchmarkConformance(b *testing.B) {
	drivertest.RunConformanceBenchmarks(b, newHarness)
}
