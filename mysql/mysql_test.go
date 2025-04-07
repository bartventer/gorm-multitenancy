package mysql

import (
	"context"
	"testing"

	"github.com/bartventer/gorm-multitenancy/mysql/v8/internal/testutil"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/drivertest"
	"gorm.io/gorm"
)

type harness struct {
	adapter *mysqlAdapter
	db      *gorm.DB
}

// Close implements [drivertest.Harness].
func (h *harness) Close() {
	// cleanup is done in MakeAdapter
}

// MakeAdapter implements [drivertest.Harness].
func (h *harness) MakeAdapter(context.Context) (adapter driver.DBFactory, tx *gorm.DB, err error) {
	return h.adapter, h.db, nil
}

// Options implements [drivertest.Harness].
func (h *harness) Options() drivertest.Options {
	return drivertest.Options{
		MaxConnectionsSQL: "SELECT @@max_connections",
	}
}

func newHarness[TB testing.TB](ctx context.Context, t TB) (drivertest.Harness, error) {
	db := testutil.NewDB(t, ctx, Open)
	return &harness{
		adapter: &mysqlAdapter{},
		db:      db,
	}, nil
}

var _ drivertest.Harness = new(harness)

func TestMySQLConformance(t *testing.T) {
	drivertest.RunConformanceTests(t, newHarness)
}
