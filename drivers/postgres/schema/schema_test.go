package schema_test

import (
	"fmt"
	"testing"

	pgschema "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema"
	"github.com/bartventer/gorm-multitenancy/v6/internal/testutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestGetSchemaNameFromDb(t *testing.T) {
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test with schema",
			args: args{
				tx: &gorm.DB{Statement: &gorm.Statement{TableExpr: &clause.Expr{SQL: "\"schema\".table"}}},
			},
			want:    "schema",
			wantErr: false,
		},
		{
			name: "Test without schema",
			args: args{
				tx: &gorm.DB{Statement: &gorm.Statement{TableExpr: &clause.Expr{SQL: "table"}}},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Test with nil TableExpr",
			args: args{
				tx: &gorm.DB{Statement: &gorm.Statement{}},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pgschema.GetSchemaNameFromDb(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSchemaNameFromDb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSchemaNameFromDb() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
