//go:build gorm_multitenancy_benchmarks
// +build gorm_multitenancy_benchmarks

package drivertest

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/internal/testmodels"
	"github.com/stretchr/testify/require"
)

// RunConformanceBenchmarks runs conformance benchmarks for driver implementations of the [driver.DBFactory] interface.
func RunConformanceBenchmarks(b *testing.B, newHarness HarnessMaker[*testing.B]) {
	b.Helper()

	b.Run("RegisterModels", func(b *testing.B) { withDB(b, newHarness, benchmarkRegisterModels) })
	b.Run("MigrateSharedModels", func(b *testing.B) { withDB(b, newHarness, benchmarkMigrateSharedModels) })
	b.Run("MigrateTenantModels", func(b *testing.B) { withDB(b, newHarness, benchmarkMigrateTenantModels) })
	b.Run("OffboardTenant", func(b *testing.B) { withDB(b, newHarness, benchmarkOffboardTenant) })
	b.Run("UseTenant", func(b *testing.B) { withDB(b, newHarness, benchmarkUseTenant) })

	// // Query Performance with Tenant Scopes
	b.Run("UseTenantCreate", func(b *testing.B) { withDB(b, newHarness, benchmarkUseTenantCreate) })
	b.Run("UseTenantFind", func(b *testing.B) { withDB(b, newHarness, benchmarkUseTenantFind) })
	b.Run("UseTenantUpdate", func(b *testing.B) { withDB(b, newHarness, benchmarkUseTenantUpdate) })
	b.Run("UseTenantDelete", func(b *testing.B) { withDB(b, newHarness, benchmarkUseTenantDelete) })
}

// benchmarkRegisterModels benchmarks the RegisterModels method.
func benchmarkRegisterModels(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	models := testmodels.MakeAllModels(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := db.RegisterModels(context.Background(), models...)
			require.NoError(b, err)
		}
	})
}

// benchmarkMigrateSharedModels benchmarks the MigrateSharedModels method.
func benchmarkMigrateSharedModels(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	models := testmodels.MakeSharedModels(b)
	err := db.RegisterModels(context.Background(), models...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = db.MigrateSharedModels(context.Background())
			require.NoError(b, err)
		}
	})
}

// benchmarkMigrateTenantModels benchmarks the MigrateTenantModels method.
func benchmarkMigrateTenantModels(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	models := testmodels.MakeAllModels(b)
	err := db.RegisterModels(context.Background(), models...)
	require.NoError(b, err)

	err = db.MigrateSharedModels(context.Background())
	require.NoError(b, err)

	var nextID atomic.Uint32

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := nextID.Add(1)
			tenant := &testmodels.Tenant{ID: fmt.Sprintf("tenant%d", id)}
			err := db.FirstOrCreate(tenant).Error
			require.NoError(b, err)
			err = db.MigrateTenantModels(context.Background(), tenant.ID)
			require.NoError(b, err)
		}
	})
}

// benchmarkOffboardTenant benchmarks the OffboardTenant method.
func benchmarkOffboardTenant(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(b, db, tenant)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := db.MigrateTenantModels(context.Background(), tenant.ID)
			require.NoError(b, err)

			err = db.OffboardTenant(context.Background(), tenant.ID)
			require.NoError(b, err)
		}
	})
}

// benchmarkUseTenant benchmarks the UseTenant method.
func benchmarkUseTenant(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(b, db, tenant)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reset, err := db.UseTenant(context.Background(), tenant.ID)
			require.NoError(b, err)

			err = reset()
			require.NoError(b, err)
		}
	})
}

// benchmarkUseTenantCreate benchmarks the Create method with tenant scoping.
func benchmarkUseTenantCreate(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	var nextID atomic.Uint32
	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(b, db, tenant)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reset, err := db.UseTenant(context.Background(), tenant.ID)
			require.NoError(b, err)
			defer reset()

			id := nextID.Add(1)
			author := &testmodels.Author{
				Tenant: *tenant,
				Books:  generateBooks(b, id),
			}
			err = db.Create(author).Error
			require.NoError(b, err)
		}
	})
}

// benchmarkUseTenantFind benchmarks the Find method with tenant scoping.
func benchmarkUseTenantFind(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(b, db, tenant)

	author := &testmodels.Author{
		Tenant: *tenant,
		Books:  generateBooks(b, 1),
	}
	err := db.Create(author).Error
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reset, err := db.UseTenant(context.Background(), tenant.ID)
			require.NoError(b, err)
			defer reset()

			var authors []testmodels.Author
			err = db.Find(&authors).Error
			require.NoError(b, err)
		}
	})
}

// benchmarkUseTenantUpdate benchmarks the Update method with tenant scoping.
func benchmarkUseTenantUpdate(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(b, db, tenant)

	author := &testmodels.Author{
		Tenant: *tenant,
		Books:  generateBooks(b, 1),
	}
	err := db.Create(author).Error
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reset, err := db.UseTenant(context.Background(), tenant.ID)
			require.NoError(b, err)
			defer reset()

			err = db.Model(author).Where("id = ?", author.ID).Update("name", "new name").Error
			require.NoError(b, err)
		}
	})
}

// benchmarkUseTenantDelete benchmarks the Delete method with tenant scoping.
func benchmarkUseTenantDelete(b *testing.B, db *multitenancy.DB, _ Options) {
	b.Helper()

	var nextID atomic.Uint32
	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(b, db, tenant)

	author := &testmodels.Author{
		Tenant: *tenant,
	}
	err := db.Create(author).Error
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reset, err := db.UseTenant(context.Background(), tenant.ID)
			require.NoError(b, err)
			defer reset()

			id := nextID.Add(1)
			book := &testmodels.Book{Title: fmt.Sprintf("Book %d", id)}
			err = db.Create(book).Error
			require.NoError(b, err)

			err = db.Where("id = ?", book.ID).Delete(book).Error
			require.NoError(b, err)
		}
	})
}

// generateBooks creates a slice of bookPrivate pointers for a given author ID.
func generateBooks(tb testing.TB, id uint32) []*testmodels.Book {
	tb.Helper()
	return []*testmodels.Book{
		{Title: fmt.Sprintf("Book 1-%d", id), Languages: []*testmodels.Language{{Name: "English"}}},
		{Title: fmt.Sprintf("Book 2-%d", id), Languages: []*testmodels.Language{{Name: "French"}}},
	}
}
