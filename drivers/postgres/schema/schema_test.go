package schema_test

import (
	"fmt"
	"testing"

	"github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/internal/testutil"
	pgschema "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/schema"
)

func TestSetSearchPath(t *testing.T) {
	// Connect to the test database.
	db := testutil.NewTestDB(testutil.WithDBName("tenants1"))

	schema := "domain1"
	// create a new schema if it does not exist
	err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error
	if err != nil {
		t.Errorf("Create schema failed, expected %v, got %v", nil, err)
	}
	defer db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))

	// Test SetSearchPath with a valid schema name.
	db, reset := pgschema.SetSearchPath(db, schema)
	if err = db.Error; err != nil {
		t.Errorf("SetSearchPath() with valid schema name failed, expected %v, got %v", nil, err)
	}

	// Test the returned ResetSearchPath function.
	err = reset()
	if err != nil {
		t.Errorf("ResetSearchPath() failed, expected %v, got %v", nil, err)
	}

	// Test SetSearchPath with an empty schema name.
	db, _ = pgschema.SetSearchPath(db, "")
	if err = db.Error; err == nil {
		t.Errorf("SetSearchPath() with empty schema name did not fail, expected an error, got %v", err)
	}
}
