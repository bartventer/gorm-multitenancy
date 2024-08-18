package drivertest

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/namespace"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
)

type mockApater struct {
	registry        *driver.ModelRegistry
	mu              sync.RWMutex
	CurrentTenantID string
}

var _ multitenancy.Adapter = new(mockApater)
var _ driver.DBFactory = new(mockApater)

// OpenDBURL implements multitenancy.Adapter.
func (m *mockApater) OpenDBURL(ctx context.Context, u *driver.URL, opts ...gorm.Option) (*multitenancy.DB, error) {
	opts = append(opts, &gorm.Config{
		DryRun: true,
	})
	db, err := gorm.Open(tests.DummyDialector{}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return multitenancy.NewDB(&mockApater{
		registry: m.registry,
	}, db), nil
}

// CurrentTenant implements [driver.DBFactory].
func (m *mockApater) CurrentTenant(ctx context.Context, db *gorm.DB) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cmp.Or(m.CurrentTenantID, "public")
}

// MigrateSharedModels implements [driver.DBFactory].
func (m *mockApater) MigrateSharedModels(ctx context.Context, db *gorm.DB) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.registry.SharedModels) == 0 {
		return errors.New("no shared models registered")
	}
	return nil
}

// MigrateTenantModels implements [driver.DBFactory].
func (m *mockApater) MigrateTenantModels(ctx context.Context, db *gorm.DB, tenantID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.registry.TenantModels) == 0 {
		return errors.New("no tenant models registered")
	}
	return nil
}

// OffboardTenant implements [driver.DBFactory].
func (m *mockApater) OffboardTenant(ctx context.Context, db *gorm.DB, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

// RegisterModels implements [driver.DBFactory].
func (m *mockApater) RegisterModels(ctx context.Context, db *gorm.DB, models ...driver.TenantTabler) error {
	gmtConfig, err := driver.NewModelRegistry(models...)
	if err != nil {
		return fmt.Errorf("failed to register models: %w", err)
	}
	m.mu.Lock()
	m.registry = gmtConfig
	m.mu.Unlock()
	return nil
}

// UseTenant implements [driver.DBFactory].
func (m *mockApater) UseTenant(ctx context.Context, db *gorm.DB, tenantID string) (reset func() error, err error) {
	if err := namespace.Validate(tenantID); err != nil {
		db.AddError(err)
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}
	m.mu.Lock()
	m.CurrentTenantID = tenantID
	m.mu.Unlock()
	return func() error {
		m.mu.Lock()
		m.CurrentTenantID = "public"
		m.mu.Unlock()
		return nil
	}, nil
}

// AdaptDB implements [multitenancy.AdapterOpener].
func (m *mockApater) AdaptDB(ctx context.Context, db *gorm.DB) (*multitenancy.DB, error) {
	return multitenancy.NewDB(&mockApater{
		registry: m.registry,
	}, db), nil
}

type harness struct {
	adpater *mockApater
	db      *gorm.DB
}

// Close implements [drivertest.Harness].
func (h *harness) Close() {
	// cleanup is done in MakeAdapter
}

// MakeAdapter implements [drivertest.Harness].
func (h *harness) MakeAdapter(context.Context) (adapter driver.DBFactory, tx *gorm.DB, err error) {
	return h.adpater, h.db, nil
}

// Options implements [drivertest.Harness].
func (h *harness) Options() Options {
	return Options{
		IsMock:            true,
		MaxConnectionsSQL: "",
	}
}

func newHarness[TB testing.TB](ctx context.Context, t TB) (Harness, error) {
	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return &harness{
		db: db,
		adpater: &mockApater{
			registry: &driver.ModelRegistry{
				SharedModels: make([]driver.TenantTabler, 0),
				TenantModels: make([]driver.TenantTabler, 0),
			},
			CurrentTenantID: "public",
		},
	}, nil
}

var _ Harness = new(harness)

func TestRunConformanceTests(t *testing.T) {
	t.Parallel()
	RunConformanceTests(t, newHarness)
}
