package drivertest

import (
	"context"
	"fmt"
	"sync"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/internal/testmodels"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/scopes"
	"github.com/stretchr/testify/require"
)

// testTenantIsolation1 tests data isolation across multiple tenants.
// It concurrently migrates tenant models and creates data for each tenant.
// It then verifies that the data is isolated to the respective tenants.
func testTenantIsolation1(t *testing.T, db *multitenancy.DB, opts Options) {
	if opts.IsMock {
		t.Skip("skipping test for mock implementations")
	}
	h := tenantHarness{TenantCount: 5, AuthorCount: 5, BookCount: 5}
	h.SetMaxConnections(t, db, opts) // bumps TenantCount
	ctx := context.TODO()

	require.NoError(t, db.RegisterModels(ctx, testmodels.MakeAllModels(t)...))
	require.NoError(t, db.MigrateSharedModels(ctx))

	tenants := h.CreateTenants(t, db)

	var wg sync.WaitGroup
	errCh := make(chan error, h.TenantCount)
	for i := range tenants {
		tenant := tenants[i]
		wg.Add(1)
		go func(tenant *testmodels.Tenant) {
			defer wg.Done()
			tenantCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			tx := db.WithContext(tenantCtx)
			if err := tx.MigrateTenantModels(tenantCtx, tenant.ID); err != nil {
				errCh <- err
				return
			}
			tx = tx.Begin()
			defer func() {
				if tx.Error == nil {
					errCh <- tx.Commit().Error
				} else {
					errCh <- tx.Rollback().Error
				}
			}()
			reset, err := tx.UseTenant(tenantCtx, tenant.ID)
			if err != nil {
				errCh <- err
				return
			}
			defer reset()
			if err := h.CreateAuthorsForTenant(tx, tenant); err != nil {
				errCh <- err
				return
			}
		}(tenant)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	// Verify data isolation.
	for i := range tenants {
		tenant := tenants[i]
		var authorCount int64
		tx := db.Model(&testmodels.Author{}).
			Scopes(scopes.WithTenantSchema(tenant.ID)).
			Where("tenant_id = ?", tenant.ID).
			Count(&authorCount)
		require.NoError(t, tx.Error)
		require.Equal(t, h.AuthorCount, int(authorCount))
	}
}

// testTenantIsolation2 tests data isolation across multiple tenants.
// It concurrently creates multiple tenants, migrates tenant models, and creates data for each tenant.
// It then verifies that the data is isolated to the respective tenants.
func testTenantIsolation2(t *testing.T, db *multitenancy.DB, opts Options) {
	if opts.IsMock {
		t.Skip("skipping test for mock implementations")
	}
	h := tenantHarness{TenantCount: 5, AuthorCount: 5, BookCount: 5}
	h.SetMaxConnections(t, db, opts) // bumps TenantCount
	ctx := context.TODO()

	require.NoError(t, db.RegisterModels(ctx, testmodels.MakeAllModels(t)...))
	require.NoError(t, db.MigrateSharedModels(ctx))

	var wg sync.WaitGroup
	errCh := make(chan error, h.TenantCount)
	for i := range h.TenantCount {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tenantCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			tx := db.WithContext(tenantCtx)
			tenant := &testmodels.Tenant{
				ID: fmt.Sprintf("tenant%d", i),
			}
			if err := tx.Create(tenant).Error; err != nil {
				errCh <- err
				return
			}
			if err := tx.MigrateTenantModels(tenantCtx, tenant.ID); err != nil {
				errCh <- err
				return
			}
			if err := tx.WithTenant(tenantCtx, tenant.ID, func(tx *multitenancy.DB) error {
				if err := h.CreateAuthorsForTenant(tx, tenant); err != nil {
					return err
				}
				return nil
			}); err != nil {
				errCh <- err
				return
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	// Verify data isolation.
	for i := range h.TenantCount {
		tenant := &testmodels.Tenant{
			ID: fmt.Sprintf("tenant%d", i),
		}
		var authorCount int64
		tx := db.Model(&testmodels.Author{}).
			Scopes(scopes.WithTenantSchema(tenant.ID)).
			Where("tenant_id = ?", tenant.ID).
			Count(&authorCount)
		require.NoError(t, tx.Error)
		require.Equal(t, h.AuthorCount, int(authorCount))
	}
}
