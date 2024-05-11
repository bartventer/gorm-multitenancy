package postgres_test

import (
	"fmt"
	"testing"

	pgdriver "github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
	"github.com/bartventer/gorm-multitenancy/v5/internal/testutil"
	pgschema "github.com/bartventer/gorm-multitenancy/v5/schema/postgres"
	scopes "github.com/bartventer/gorm-multitenancy/v5/scopes"
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

type Tenant struct {
	pgdriver.TenantPKModel
}

func (Tenant) TableName() string { return "public.tenants" }

type Book struct {
	gorm.Model
	TenantID string `gorm:"column:tenant_schema"`
	Tenant   Tenant `gorm:"foreignKey:TenantID;references:SchemaName"`
}

func (Book) TableName() string   { return "books" }
func (Book) IsTenantTable() bool { return true }

func BenchmarkSetSearchPath(b *testing.B) {
	db, err := gorm.Open(pgdriver.Open(testutil.GetDSN(testutil.WithDBName("tenants2"))), &gorm.Config{})
	if err != nil {
		b.Fatalf("Open failed: %v", err)
	}
	if err := pgdriver.RegisterModels(db, &Tenant{}, &Book{}); err != nil {
		b.Fatalf("RegisterModels failed: %v", err)
	}
	if err := pgdriver.MigratePublicSchema(db); err != nil {
		b.Fatalf("Migrate public schema failed: %v", err)
	}
	tenant := Tenant{
		TenantPKModel: pgdriver.TenantPKModel{
			DomainURL:  "tenant1.example.com",
			SchemaName: "tenant1",
		},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tenant).Error; err != nil {
		b.Fatalf("Create tenant failed: %v", err)
	}
	if err := pgdriver.CreateSchemaForTenant(db, tenant.SchemaName); err != nil {
		b.Fatalf("Create schema for tenant failed: %v", err)
	}
	defer pgdriver.DropSchemaForTenant(db, tenant.SchemaName)

	b.ResetTimer()
	b.Run("Create with WithTenantSchema", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			book := Book{
				Tenant: tenant,
			}
			if err := db.Scopes(scopes.WithTenantSchema(tenant.SchemaName)).Create(&book).Error; err != nil {
				b.Errorf("Create book failed, expected %v, got %v", nil, err)
			}
		}
	})
	b.Run("Create with SetSearchPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, resetSearchPath := pgschema.SetSearchPath(db, tenant.SchemaName)
			if err := db.Error; err != nil {
				b.Errorf("SetSearchPath() with valid schema name failed, expected %v, got %v", nil, err)
			}
			defer resetSearchPath()
			book := Book{
				Tenant: tenant,
			}
			if err := db.Create(&book).Error; err != nil {
				b.Errorf("Create book failed, expected %v, got %v", nil, err)
			}
		}
	})
}
