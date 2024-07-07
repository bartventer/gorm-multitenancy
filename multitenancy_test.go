package multitenancy

import (
	"context"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
)

type mockDriver struct{}

func (m *mockDriver) CurrentTenant(ctx context.Context, db *gorm.DB) string {
	return "test-tenant"
}

func (m *mockDriver) RegisterModels(ctx context.Context, db *gorm.DB, models ...driver.TenantTabler) error {
	return nil
}

func (m *mockDriver) MigrateSharedModels(ctx context.Context, db *gorm.DB) error {
	return nil
}

func (m *mockDriver) MigrateTenantModels(ctx context.Context, db *gorm.DB, tenantID string) error {
	return nil
}

func (m *mockDriver) OffboardTenant(ctx context.Context, db *gorm.DB, tenantID string) error {
	return nil
}

func (m *mockDriver) UseTenant(ctx context.Context, db *gorm.DB, tenantID string) (reset func() error, err error) {
	return func() error { return nil }, nil
}

func TestDB_CurrentTenant(t *testing.T) {
	db := NewDB(&mockDriver{}, &gorm.DB{})
	tenantID := db.CurrentTenant(context.Background())
	if tenantID != "test-tenant" {
		t.Errorf("Expected tenant ID 'test-tenant', got '%s'", tenantID)
	}
}

func TestDB_RegisterModels(t *testing.T) {
	db := NewDB(&mockDriver{}, &gorm.DB{})
	err := db.RegisterModels(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDB_MigrateSharedModels(t *testing.T) {
	db := NewDB(&mockDriver{}, &gorm.DB{})
	err := db.MigrateSharedModels(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDB_MigrateTenantModels(t *testing.T) {
	db := NewDB(&mockDriver{}, &gorm.DB{})
	err := db.MigrateTenantModels(context.Background(), "test-tenant")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDB_OffboardTenant(t *testing.T) {
	db := NewDB(&mockDriver{}, &gorm.DB{})
	err := db.OffboardTenant(context.Background(), "test-tenant")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDB_UseTenant(t *testing.T) {
	db := NewDB(&mockDriver{}, &gorm.DB{})
	_, err := db.UseTenant(context.Background(), "test-tenant")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func newDB(t *testing.T) *DB {
	t.Helper()
	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{})
	require.NoError(t, err)
	return NewDB(&mockDriver{}, db)
}

func TestDB_Session(t *testing.T) {
	db := newDB(t)
	newDB := db.Session(&gorm.Session{})
	assert.NotNil(t, newDB)
}

func TestDB_Debug(t *testing.T) {
	db := newDB(t)
	newDB := db.Debug()
	assert.NotNil(t, newDB)
}

func TestDB_WithContext(t *testing.T) {
	db := newDB(t)
	newDB := db.WithContext(context.Background())
	assert.NotNil(t, newDB)
}

func TestDB_Begin(t *testing.T) {
	db := newDB(t)
	newDB := db.Begin()
	assert.NotNil(t, newDB)
}
