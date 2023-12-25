package postgres

import (
	"errors"
	"fmt"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy"
	"github.com/bartventer/gorm-multitenancy/internal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type testPublicTable struct {
	gorm.Model
	TenantModel
	SubdomainURL string
}

func (testPublicTable) TableName() string {
	return "test_public_table"
}

type testTenantTable struct {
	gorm.Model
	PrivateField string
}

func (testTenantTable) TableName() string {
	return "test_private_table"
}

var _ multitenancy.TenantTabler = (*testTenantTable)(nil)

func (testTenantTable) IsTenantTable() bool { return true }

var (
	testDb = internal.NewTestDB()

	testDbWithError = internal.NewTestDB().Scopes(func(d *gorm.DB) *gorm.DB {
		d.AddError(errors.New("invalid db"))
		return d
	})
)

func TestMain(m *testing.M) {
	m.Run()

	// drop public tables
	fmt.Println("[multitenancy] ⏳ tearing down... dropping public tables")
	testDb.Exec(fmt.Sprintf("SET search_path TO %s", "public"))
	testDb.Migrator().DropTable(
		&testPublicTable{},
		&testTenantTable{},
	)
	fmt.Println("[multitenancy] ✅ teardown complete")
}

func TestMigrator_CreateSchemaForTenant(t *testing.T) {
	type fields struct {
		Migrator           postgres.Migrator
		multitenancyConfig *multitenancyConfig
	}
	type args struct {
		tenant string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				Migrator: testDb.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testPublicTable{}, &testTenantTable{}},
				},
			},
			args: args{
				tenant: "test_tenant",
			},
			wantErr: false,
		},
		{
			name: "invalid",
			fields: fields{
				Migrator: testDbWithError.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{},
					tenantModels: []interface{}{},
					models:       []interface{}{},
				},
			},
			args: args{
				tenant: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid schema name (reserved)",
			fields: fields{
				Migrator: testDb.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testPublicTable{}, &testTenantTable{}},
				},
			},
			args: args{
				tenant: "pg_tenant1",
			},
			wantErr: true,
		},
		{
			name: "invalid: no private tables to migrate",
			fields: fields{
				Migrator: testDb.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					models:       []interface{}{&testPublicTable{}},
				},
			},
			args: args{
				tenant: "test_tenant",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migrator{
				Migrator:           tt.fields.Migrator,
				multitenancyConfig: tt.fields.multitenancyConfig,
			}
			if err := m.CreateSchemaForTenant(tt.args.tenant); (err != nil) != tt.wantErr {
				t.Errorf("Migrator.CreateSchemaForTenant() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Cleanup(func() {
				if !tt.wantErr { // cleanup; drop schema if test passed
					testDb.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tt.args.tenant))
				}
			})
		})
	}
}

func TestMigrator_MigratePublicSchema(t *testing.T) {
	type fields struct {
		Migrator           postgres.Migrator
		multitenancyConfig *multitenancyConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				Migrator: testDb.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testPublicTable{}, &testTenantTable{}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid",
			fields: fields{
				Migrator: testDbWithError.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{},
					tenantModels: []interface{}{},
					models:       []interface{}{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid: no public tables to migrate",
			fields: fields{
				Migrator: testDb.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testTenantTable{}},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migrator{
				Migrator:           tt.fields.Migrator,
				multitenancyConfig: tt.fields.multitenancyConfig,
			}
			if err := m.MigratePublicSchema(); (err != nil) != tt.wantErr {
				t.Errorf("Migrator.MigratePublicSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Cleanup(func() {
				if !tt.wantErr { // cleanup; drop schema if test passed
					// drop table
					testDb.Migrator().DropTable(
						&testPublicTable{},
						&testTenantTable{},
					)
				}
			})
		})
	}
}

func TestMigrator_DropSchemaForTenant(t *testing.T) {
	type fields struct {
		Migrator           postgres.Migrator
		multitenancyConfig *multitenancyConfig
	}
	type args struct {
		tenant string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				Migrator: testDb.Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testPublicTable{}, &testTenantTable{}},
				},
			},
			args: args{
				tenant: "test_tenant",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migrator{
				Migrator:           tt.fields.Migrator,
				multitenancyConfig: tt.fields.multitenancyConfig,
			}
			// create schema
			if err := m.DB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tt.args.tenant)).Error; err != nil {
				t.Errorf("setSearchPath() error = %v, wantErr %v", err, false)
			}
			if err := m.DropSchemaForTenant(tt.args.tenant); (err != nil) != tt.wantErr {
				t.Errorf("Migrator.DropSchemaForTenant() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ensure schema does not exist
			var exists bool
			if err := m.DB.Raw(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = '%s')", tt.args.tenant)).Scan(&exists).Error; err != nil {
				t.Errorf("setSearchPath() error = %v, wantErr %v", err, false)
			}
			if exists {
				t.Errorf("setSearchPath() schema %s still exists", tt.args.tenant)
			}
		})
	}
}

func Test_setSearchPath(t *testing.T) {
	type args struct {
		db     *gorm.DB
		schema string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				db:     testDb,
				schema: "public",
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				db: testDbWithError,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setSearchPath(tt.args.db, tt.args.schema); (err != nil) != tt.wantErr {
				t.Errorf("setSearchPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigrator_AutoMigrate(t *testing.T) {
	type fields struct {
		Migrator           postgres.Migrator
		multitenancyConfig *multitenancyConfig
	}
	type args struct {
		values []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid: with valid migrate option",
			fields: fields{
				Migrator: testDb.Scopes(withMigrationOption(multiMigrationOptionMigratePublicTables)).Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testPublicTable{}, &testTenantTable{}},
				},
			},
			args: args{
				values: []interface{}{
					&testPublicTable{},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: with invalid migrate option",
			fields: fields{
				Migrator: testDb.Scopes(withMigrationOption(multitenancyMigrationOption(0))).Migrator().(postgres.Migrator),
				multitenancyConfig: &multitenancyConfig{
					publicModels: []interface{}{&testPublicTable{}},
					tenantModels: []interface{}{&testTenantTable{}},
					models:       []interface{}{&testPublicTable{}, &testTenantTable{}},
				},
			},
			args: args{
				values: []interface{}{
					&testPublicTable{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Migrator{
				Migrator:           tt.fields.Migrator,
				multitenancyConfig: tt.fields.multitenancyConfig,
			}
			if err := m.AutoMigrate(tt.args.values...); (err != nil) != tt.wantErr {
				t.Errorf("Migrator.AutoMigrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
