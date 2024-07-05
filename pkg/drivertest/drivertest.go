// Package drivertest provides conformance tests for db implementations.
package drivertest

import (
	"context"
	"strings"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v7"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type (
	// Options is a set of options to configure driver specific behavior.
	// This should be used as a last resort to for example skip tests for a feature
	// that is not supported by the driver. Not encouraged.
	Options struct {
		// IsMock indicates if the driver is a mock implementation.
		// This is useful for mock implementations or drivers that do not support
		// the specific features required by the tests.
		IsMock bool
	}

	// Harness descibes the functionality test harnesses must provide to run
	// conformance tests.
	Harness interface {
		// MakeAdapter makes a new adapter for the harness.
		// Adapter and tx should be considered valid if err is nil.
		MakeAdapter(context.Context) (adapter driver.DBFactory, tx *gorm.DB, err error)

		// Close closes resources used by the harness.
		Close()

		// Options returns the options for the harness.
		Options() Options
	}

	// HarnessMaker describes functions that construct a harness for running tests.
	// It is called exactly once per test; Harness.Close() will be called when the test is complete.
	HarnessMaker[TB testing.TB] func(ctx context.Context, t TB) (Harness, error)
)

// RunConformanceTests runs conformance tests for driver implementations of the [driver.DBFactory] interface.
func RunConformanceTests(t *testing.T, newHarness HarnessMaker[*testing.T]) {
	t.Helper()

	t.Run("RegisterModels", func(t *testing.T) { withDB(t, newHarness, testRegisterModels) })
	t.Run("MigrateSharedModels", func(t *testing.T) { withDB(t, newHarness, testMigrateSharedModels) })
	t.Run("MigrateTenantModels", func(t *testing.T) { withDB(t, newHarness, testMigrateTenantModels) })
	t.Run("OffboardTenant", func(t *testing.T) { withDB(t, newHarness, testOffboardTenant) })
	t.Run("UseTenant", func(t *testing.T) { withDB(t, newHarness, testUseTenant) })
	t.Run("CurrentTenant", func(t *testing.T) { withDB(t, newHarness, testCurrentTenant) })
	t.Run("TenantModel", func(t *testing.T) { withDB(t, newHarness, testTenantModel) })
	t.Run("InvalidAutoMigrate", func(t *testing.T) { withDB(t, newHarness, testInvalidAutoMigrate) })
}

// withDB creates a new DB and runs the test function.
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

	f(t, db, h.Options())
}

// testRegisterModels tests the RegisterModels method.
func testRegisterModels(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Parallel()

	t.Run("valid models", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), makeAllModels(t)...)
		assert.NoError(t, err)
	})

	t.Run("invalid shared model", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), &userSharedInvalid{})
		assert.Error(t, err, "expected error, got nil")
	})

	t.Run("invalid tenant model", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), &bookPrivateInvalid{})
		assert.Error(t, err, "expected error, got nil")
	})
}

// testMigrateSharedModels tests the MigrateSharedModels method.
func testMigrateSharedModels(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Parallel()

	t.Run("no public models", func(t *testing.T) {
		err := db.MigrateSharedModels(context.Background())
		assert.Error(t, err, "expected error, got nil")
	})

	t.Run("valid public models", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), makeSharedModels(t)...)
		require.NoError(t, err)

		err = db.MigrateSharedModels(context.Background())
		assert.NoError(t, err)
	})
}

// testMigrateTenantModels tests the MigrateTenantModels method.
func testMigrateTenantModels(t *testing.T, db *multitenancy.DB, opts Options) {
	t.Parallel()

	err := db.RegisterModels(context.Background(), makeSharedModels(t)...)
	require.NoError(t, err)

	err = db.MigrateSharedModels(context.Background())
	require.NoError(t, err)

	tenant := &userShared{ID: "tenant1"}
	err = db.FirstOrCreate(tenant).Error
	require.NoError(t, err)

	t.Run("no tenant models", func(t *testing.T) {
		err := db.MigrateTenantModels(context.Background(), tenant.ID)
		assert.Error(t, err, "expected error, got nil")
	})

	t.Run("valid tenant models", func(t *testing.T) {
		ctx := context.Background()
		err := db.RegisterModels(ctx, makePrivateModels(t)...)
		require.NoError(t, err)

		err = db.MigrateTenantModels(ctx, tenant.ID)
		assert.NoError(t, err)
	})

	t.Run("existing transaction", func(t *testing.T) {
		if opts.IsMock {
			t.Skip("skipping transaction test for mock implementations")
		}
		err := db.RegisterModels(context.Background(), makePrivateModels(t)...)
		require.NoError(t, err)
		err = db.Transaction(func(tx *multitenancy.DB) error {
			return tx.MigrateTenantModels(context.Background(), tenant.ID)
		})
		assert.NoError(t, err)
	})
}

// testOffboardTenant tests the OffboardTenant method.
func testOffboardTenant(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Parallel()

	tenant := &userShared{ID: "tenant1"}
	setupTenant(t, db, tenant)

	err := db.OffboardTenant(context.Background(), tenant.ID)
	assert.NoError(t, err)
}

// testUseTenant tests the UseTenant method.
func testUseTenant(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Parallel()

	tenant := &userShared{ID: "tenant1"}
	setupTenant(t, db, tenant)

	t.Run("TestCRUD", func(t *testing.T) {
		ctx := context.Background()
		reset, err := db.UseTenant(ctx, tenant.ID)
		require.NoError(t, err)
		defer reset()

		author := &authorPrivate{
			User: *tenant,
			Books: []*bookPrivate{
				{Title: "Book 1", Languages: []*languagePrivate{{Name: "English"}}},
				{Title: "Book 2", Languages: []*languagePrivate{{Name: "French"}}},
			},
		}
		err = db.Create(author).Error
		assert.NoError(t, err)
	})

	// Check reset
	t.Run("TestReset", func(t *testing.T) {
		ctx := context.Background()
		reset, err := db.UseTenant(ctx, tenant.ID)
		require.NoError(t, err)

		err = reset()
		assert.NoError(t, err)
	})

	// Check invalid empty tenant
	t.Run("TestReset", func(t *testing.T) {
		ctx := context.Background()
		_, err := db.UseTenant(ctx, "")
		require.Error(t, err)
	})
}

// testCurrentTenant tests the CurrentTenant method.
func testCurrentTenant(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Parallel()

	ctx := context.Background()
	assert.Equal(t, "public", db.CurrentTenant(ctx), "expected initial tenant context")

	tenant := &userShared{ID: "tenant1"}
	setupTenant(t, db, tenant)

	reset, err := db.UseTenant(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, db.CurrentTenant(ctx), "expected tenant context")
	err = reset()
	require.NoError(t, err)

	assert.Equal(t, "public", db.CurrentTenant(ctx), "expected tenant context after reset")
}

// testTenantModel tests the TenantModel struct.
func testTenantModel(t *testing.T, db *multitenancy.DB, opts Options) {
	t.Parallel()
	ctx := context.Background()

	err := db.RegisterModels(ctx, mockTenantModel{})
	require.NoError(t, err)

	err = db.MigrateSharedModels(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		data    *multitenancy.TenantModel
		wantErr bool
	}{
		{
			name: "valid: check constraints",
			data: &multitenancy.TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "tenant1",
			},
			wantErr: false,
		},
		{
			name: "invalid: ID too long",
			data: &multitenancy.TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: strings.Repeat("a", 64),
			},
			wantErr: true,
		},
		{
			name: "invalid: ID too short",
			data: &multitenancy.TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "a",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if opts.IsMock {
				t.Skip("skipping test for mock implementations")
			}
			m := &mockTenantModel{
				TenantModel: *tt.data,
			}
			err := db.Create(m).Error
			assert.Equal(t, tt.wantErr, (err != nil), "expected error")
			if err != nil {
				return
			}
			t.Cleanup(func() {
				err := db.Unscoped().Delete(m).Error
				require.NoError(t, err)
			})
		})
	}
}

// testInvalidAutoMigrate tests the invalid auto migration.
func testInvalidAutoMigrate(t *testing.T, db *multitenancy.DB, opts Options) {
	t.Parallel()
	if opts.IsMock {
		t.Skip("skipping test for mock implementations; not supported")
	}

	err := db.AutoMigrate(&userShared{})
	assert.ErrorIs(t, err, driver.ErrInvalidMigration)
}

// setupTenant sets up a tenant for testing.
// It registers models, migrates shared models, creates the tenant, and migrates tenant models.
func setupTenant[TB testing.TB](t TB, db *multitenancy.DB, tenant *userShared) {
	t.Helper()

	err := db.RegisterModels(context.Background(), makeAllModels(t)...)
	require.NoError(t, err)

	err = db.MigrateSharedModels(context.Background())
	require.NoError(t, err)

	err = db.FirstOrCreate(tenant).Error
	require.NoError(t, err)

	err = db.MigrateTenantModels(context.Background(), tenant.ID)
	require.NoError(t, err)
}
