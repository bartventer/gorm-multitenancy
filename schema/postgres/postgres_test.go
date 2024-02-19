package postgres

import (
	"fmt"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v3/internal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Test_getSchemaNameFromSqlExpr(t *testing.T) {
	type args struct {
		tableExprSQL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with schema and table",
			args: args{
				tableExprSQL: "\"test_domain\".\"mock_private\"",
			},
			want: "test_domain",
		},
		{
			name: "Test with only table",
			args: args{
				tableExprSQL: "\"mock_private\"",
			},
			want: "",
		},
		{
			name: "Test with empty string",
			args: args{
				tableExprSQL: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSchemaNameFromSQLExpr(tt.args.tableExprSQL); got != tt.want {
				t.Errorf("getSchemaNameFromTableExpreSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
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
			got, err := GetSchemaNameFromDb(tt.args.tx)
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
	db := internal.NewTestDB()

	schema := "domain1"
	// create a new schema if it does not exist
	err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error
	if err != nil {
		t.Errorf("Create schema failed, expected %v, got %v", nil, err)
	}
	defer db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))

	// Test SetSearchPath with a valid schema name.
	db, reset, err := SetSearchPath(db, schema)
	if err != nil {
		t.Errorf("SetSearchPath() with valid schema name failed, expected %v, got %v", nil, err)
	}

	// Test the returned ResetSearchPath function.
	err = reset()
	if err != nil {
		t.Errorf("ResetSearchPath() failed, expected %v, got %v", nil, err)
	}

	// Test SetSearchPath with an empty schema name.
	_, _, err = SetSearchPath(db, "")
	if err == nil {
		t.Errorf("SetSearchPath() with empty schema name did not fail, expected an error, got %v", err)
	}

}

func BenchmarkSetSearchPath(b *testing.B) {
	// Connect to the test database.
	db := internal.NewTestDB()

	schema := "domain1"
	// create a new schema if it does not exist
	err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error
	if err != nil {
		b.Errorf("Create schema failed, expected %v, got %v", nil, err)
	}
	defer db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))

	// Benchmark SetSearchPath with a valid schema name.
	b.Run("SetSearchPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := SetSearchPath(db, schema)
			if err != nil {
				b.Errorf("SetSearchPath() with valid schema name failed, expected %v, got %v", nil, err)
			}
		}
	})

	// Benchmark ResetSearchPath.
	b.Run("ResetSearchPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, reset, err := SetSearchPath(db, schema)
			if err != nil {
				b.Errorf("SetSearchPath() with valid schema name failed, expected %v, got %v", nil, err)
			}
			err = reset()
			if err != nil {
				b.Errorf("ResetSearchPath() failed, expected %v, got %v", nil, err)
			}
		}
	})
}
