package drivertest

import (
	"cmp"
	"context"
	"fmt"
	"runtime"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/internal/testmodels"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func withDB[TB testing.TB](t TB, newHarness HarnessMaker[TB], f func(TB, *multitenancy.DB, Options)) {
	t.Helper()

	ctx := context.Background()
	h, err := newHarness(ctx, t)
	require.NoError(t, err)
	defer h.Close()

	a, gdb, err := h.MakeAdapter(ctx) // adapter and gorm db
	require.NoError(t, err)
	db := multitenancy.NewDB(a, gdb) // multitenancy db
	require.NoError(t, err)

	f(t, db.Session(&gorm.Session{}), h.Options())
}

func parallel(t *testing.T, newHarness HarnessMaker[*testing.T], f func(*testing.T, *multitenancy.DB, Options)) {
	t.Helper()
	t.Parallel()
	withDB(t, newHarness, f)
}

type setupModelsOptions struct {
	SkipRegisterModels  bool
	SkipSharedMigration bool
	SkipCreateTenant    bool
	SkipTenantMigration bool
}

type setupModelsOption func(*setupModelsOptions)

// setupModels sets up a tenant for testing.
// It registers models, migrates shared models, creates the tenant, and migrates tenant models.
func setupModels[TB testing.TB](t TB, db *multitenancy.DB, tenant *testmodels.Tenant, opts ...setupModelsOption) {
	t.Helper()

	options := setupModelsOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	var err error
	if !options.SkipRegisterModels {
		err = db.RegisterModels(context.Background(), testmodels.MakeAllModels(t)...)
		require.NoError(t, err)
	}

	if !options.SkipSharedMigration {
		err = db.MigrateSharedModels(context.Background())
		require.NoError(t, err)
	}

	if !options.SkipCreateTenant {
		err = db.FirstOrCreate(tenant).Error
		require.NoError(t, err)
	}

	if !options.SkipTenantMigration {
		err = db.MigrateTenantModels(context.Background(), tenant.ID)
		require.NoError(t, err)
	}
}

type tenantHarness struct {
	TenantCount, AuthorCount, BookCount int
}

// Not intended for concurrent use.
func (h *tenantHarness) SetMaxConnections(tb testing.TB, db *multitenancy.DB, opts Options) {
	tb.Helper()

	sqlDB, err := db.DB.DB()
	require.NoError(tb, err)

	var maxConns int
	require.NoError(tb, db.Raw(opts.MaxConnectionsSQL).Scan(&maxConns).Error)

	numCores := runtime.NumCPU()
	conns := max(numCores*4, maxConns/2)
	sqlDB.SetMaxOpenConns(conns)

	tb.Logf("num cores: %d", numCores)
	tb.Logf("max connections: %d", maxConns)
	tb.Logf("set max connections to %d", conns)

	if cmp.Less(h.TenantCount, conns) {
		h.TenantCount = conns
		tb.Logf("increased tenant count to %d", h.TenantCount)
	}
}

// Not intended for concurrent use.
func (h *tenantHarness) CreateTenants(tb testing.TB, db *multitenancy.DB) []*testmodels.Tenant {
	tb.Helper()
	tenants := h.MakeTenants(tb, h.TenantCount)
	require.NoError(tb, db.Create(&tenants).Error, "failed to create tenants")
	return tenants
}

func (h *tenantHarness) MakeTenants(tb testing.TB, count int) []*testmodels.Tenant {
	tb.Helper()
	tenants := make([]*testmodels.Tenant, count)
	for i := range tenants {
		tenants[i] = &testmodels.Tenant{ID: fmt.Sprintf("tenant%d", i)}
	}
	return tenants
}

// Intended for concurrent use.
func (h *tenantHarness) CreateAuthorsForTenant(tx *multitenancy.DB, tenant *testmodels.Tenant) error {
	authors := h.MakeAuthorsForTenant(tenant, h.AuthorCount, h.BookCount)
	return tx.Create(&authors).Error
}

func (h *tenantHarness) MakeAuthorsForTenant(tenant *testmodels.Tenant, authorCount, bookCount int) []*testmodels.Author {
	authors := make([]*testmodels.Author, authorCount)
	for i := range authors {
		authors[i] = &testmodels.Author{
			Tenant: *tenant,
			Books: func(count int) []*testmodels.Book {
				books := make([]*testmodels.Book, count)
				for i := range books {
					books[i] = &testmodels.Book{
						Title: fmt.Sprintf("%s-book%d", tenant.ID, i),
						Languages: []*testmodels.Language{
							{Name: "Afrikaans"},
							{Name: "English"},
						},
					}
				}
				return books
			}(bookCount),
		}
	}
	return authors
}
