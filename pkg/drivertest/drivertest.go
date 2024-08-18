// Package drivertest provides conformance tests for db implementations.
package drivertest

import (
	"context"
	"strings"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/internal/testmodels"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type (
	// Options is a set of options to configure driver specific behavior.
	Options struct {
		// IsMock indicates if the driver is a mock implementation.
		// This is useful for mock implementations or drivers that do not support
		// the specific features required by the tests.
		IsMock bool

		// MaxConnectionsSQL specifies the SQL query to get the maximum number of connections.
		MaxConnectionsSQL string
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

	t.Run("RegisterModels", func(t *testing.T) { parallel(t, newHarness, testRegisterModels) })
	t.Run("MigrateSharedModels", func(t *testing.T) { parallel(t, newHarness, testMigrateSharedModels) })
	t.Run("MigrateTenantModels", func(t *testing.T) { parallel(t, newHarness, testMigrateTenantModels) })
	t.Run("OffboardTenant", func(t *testing.T) { parallel(t, newHarness, testOffboardTenant) })
	t.Run("UseTenant", func(t *testing.T) { parallel(t, newHarness, testUseTenant) })
	t.Run("WithTenant", func(t *testing.T) { parallel(t, newHarness, testWithTenant) })
	t.Run("CurrentTenant", func(t *testing.T) { parallel(t, newHarness, testCurrentTenant) })
	t.Run("TenantModel", func(t *testing.T) { parallel(t, newHarness, testTenantModel) })
	t.Run("DBInstance", func(t *testing.T) { parallel(t, newHarness, testDBInstance) })

	t.Run("Concurrency", func(t *testing.T) {
		t.Run("TenantIsolation1", func(t *testing.T) { parallel(t, newHarness, testTenantIsolation1) })
		t.Run("TenantIsolation2", func(t *testing.T) { parallel(t, newHarness, testTenantIsolation2) })
	})
}

// testRegisterModels tests the RegisterModels method.
func testRegisterModels(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Run("valid models", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), testmodels.MakeAllModels(t)...)
		assert.NoError(t, err)
	})

	t.Run("invalid shared model", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), &testmodels.TenantInvalid{})
		assert.Error(t, err, "expected error, got nil")
	})

	t.Run("invalid tenant model", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), &testmodels.BookInvalid{})
		assert.Error(t, err, "expected error, got nil")
	})
}

// testMigrateSharedModels tests the MigrateSharedModels method.
func testMigrateSharedModels(t *testing.T, db *multitenancy.DB, _ Options) {
	t.Run("no public models", func(t *testing.T) {
		err := db.MigrateSharedModels(context.Background())
		assert.Error(t, err, "expected error, got nil")
	})

	t.Run("valid public models", func(t *testing.T) {
		err := db.RegisterModels(context.Background(), testmodels.MakeSharedModels(t)...)
		require.NoError(t, err)

		err = db.MigrateSharedModels(context.Background())
		assert.NoError(t, err)
	})
}

// testMigrateTenantModels tests the MigrateTenantModels method.
func testMigrateTenantModels(t *testing.T, db *multitenancy.DB, opts Options) {
	err := db.RegisterModels(context.Background(), testmodels.MakeSharedModels(t)...)
	require.NoError(t, err)

	err = db.MigrateSharedModels(context.Background())
	require.NoError(t, err)

	tenant := &testmodels.Tenant{ID: "tenant1"}
	err = db.FirstOrCreate(tenant).Error
	require.NoError(t, err)

	t.Run("no tenant models", func(t *testing.T) {
		err := db.MigrateTenantModels(context.Background(), tenant.ID)
		assert.Error(t, err, "expected error, got nil")
	})

	t.Run("valid tenant models", func(t *testing.T) {
		ctx := context.Background()
		err := db.RegisterModels(ctx, testmodels.MakePrivateModels(t)...)
		require.NoError(t, err)

		err = db.MigrateTenantModels(ctx, tenant.ID)
		assert.NoError(t, err)
	})

	t.Run("existing transaction", func(t *testing.T) {
		if opts.IsMock {
			t.Skip("skipping transaction test for mock implementations")
		}
		err := db.RegisterModels(context.Background(), testmodels.MakePrivateModels(t)...)
		require.NoError(t, err)
		err = db.Transaction(func(tx *multitenancy.DB) error {
			return tx.MigrateTenantModels(context.Background(), tenant.ID)
		})
		assert.NoError(t, err)
	})
}

// testOffboardTenant tests the OffboardTenant method.
func testOffboardTenant(t *testing.T, db *multitenancy.DB, _ Options) {
	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(t, db, tenant)

	err := db.OffboardTenant(context.Background(), tenant.ID)
	assert.NoError(t, err)
}

// testUseTenant tests the UseTenant method.
func testUseTenant(t *testing.T, db *multitenancy.DB, _ Options) {
	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(t, db, tenant)

	t.Run("TestCRUD", func(t *testing.T) {
		ctx := context.Background()
		reset, err := db.UseTenant(ctx, tenant.ID)
		require.NoError(t, err)
		defer reset()

		author := &testmodels.Author{
			Tenant: *tenant,
			Books: []*testmodels.Book{
				{Title: "Book 1", Languages: []*testmodels.Language{{Name: "English"}}},
				{Title: "Book 2", Languages: []*testmodels.Language{{Name: "French"}}},
			},
		}
		err = db.Create(author).Error
		assert.NoError(t, err)
	})

	t.Run("TestReset", func(t *testing.T) {
		ctx := context.Background()
		reset, err := db.UseTenant(ctx, tenant.ID)
		require.NoError(t, err)

		err = reset()
		assert.NoError(t, err)
	})

	t.Run("TestEmptyTenant", func(t *testing.T) {
		ctx := context.Background()
		_, err := db.UseTenant(ctx, "")
		require.Error(t, err)
	})
}

// testWithTenant tests the WithTenant method.
func testWithTenant(t *testing.T, db *multitenancy.DB, opts Options) {
	if opts.IsMock {
		t.Skip("skipping test for mock implementations; not supported")
	}
	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(t, db, tenant)
	assert.Equal(t, "public", db.CurrentTenant(context.Background()), "expected initial tenant context")

	ctx := context.Background()
	err := db.WithTenant(ctx, tenant.ID, func(tx *multitenancy.DB) error {
		assert.Equal(t, tenant.ID, tx.CurrentTenant(ctx), "expected tenant context")
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "public", db.CurrentTenant(ctx), "expected initial tenant context")
}

// testCurrentTenant tests the CurrentTenant method.
func testCurrentTenant(t *testing.T, db *multitenancy.DB, _ Options) {
	ctx := context.Background()
	assert.Equal(t, "public", db.CurrentTenant(ctx), "expected initial tenant context")

	tenant := &testmodels.Tenant{ID: "tenant1"}
	setupModels(t, db, tenant)

	reset, err := db.UseTenant(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, db.CurrentTenant(ctx), "expected tenant context")
	err = reset()
	require.NoError(t, err)

	assert.Equal(t, "public", db.CurrentTenant(ctx), "expected tenant context after reset")
}

// testTenantModel tests the TenantModel struct.
func testTenantModel(t *testing.T, db *multitenancy.DB, opts Options) {
	ctx := context.Background()

	err := db.RegisterModels(ctx, testmodels.FakeTenant{})
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
			m := &testmodels.FakeTenant{
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

// testDBInstance tests the DB instance.
func testDBInstance(t *testing.T, db *multitenancy.DB, opts Options) {
	if opts.IsMock {
		t.Skip("skipping test for mock implementations; not supported")
	}

	t.Run("InvalidAutoMigrate", func(t *testing.T) {
		err := db.AutoMigrate(&testmodels.Tenant{})
		require.ErrorIs(t, err, driver.ErrInvalidMigration)
	})

	t.Run("Transaction", func(t *testing.T) {
		err := db.Transaction(func(tx *multitenancy.DB) error {
			sqlTx, ok := tx.Statement.ConnPool.(gorm.TxCommitter)
			require.True(t, ok, "expected sql.Tx")
			require.NotNil(t, sqlTx, "expected sql.Tx")
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Begin", func(t *testing.T) {
		tx := db.Begin()
		sqlTx, ok := tx.Statement.ConnPool.(gorm.TxCommitter)
		require.True(t, ok, "expected sql.Tx")
		require.NotNil(t, sqlTx, "expected sql.Tx")
	})
}
