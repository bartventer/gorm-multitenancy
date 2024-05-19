package postgres

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v6"
	"github.com/bartventer/gorm-multitenancy/v6/internal/testutil"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

type (
	invalidNonTabler struct {
		gorm.Model
		Name string
	}

	invalidPublicUser struct {
		gorm.Model
		Name string
	}

	invalidPrivateProduct struct {
		gorm.Model
		Name string
	}

	testUser struct {
		gorm.Model
		Name string
	}

	testProduct struct {
		gorm.Model
		Name string
	}
)

func (invalidPublicUser) TableName() string {
	return "users" // invalid table name; should be "public.users"
}

func (invalidPrivateProduct) TableName() string {
	return fmt.Sprintf("%s.products", PublicSchemaName) // invalid table name; should be "products"
}

var _ multitenancy.TenantTabler = (*invalidPrivateProduct)(nil)

func (invalidPrivateProduct) IsTenantTable() bool { return true }

func (testUser) TableName() string {
	return fmt.Sprintf("%s.users", PublicSchemaName)
}

func (testProduct) TableName() string {
	return "products"
}

var _ multitenancy.TenantTabler = (*testProduct)(nil)

func (testProduct) IsTenantTable() bool { return true }

var dsn = testutil.GetDSN()

func TestOpen(t *testing.T) {
	type args struct {
		dsn    string
		models []interface{}
	}
	tests := []struct {
		name string
		args args
		want gorm.Dialector
	}{
		{
			name: "Test Open",
			args: args{
				dsn:    dsn,
				models: []interface{}{&testUser{}, &testProduct{}},
			},
			want: &Dialector{
				rw:        &sync.RWMutex{},
				Dialector: *postgres.Open(dsn).(*postgres.Dialector),
				migratorConfig: func() *migratorConfig {
					cfg, _ := newMigratorConfig([]interface{}{&testUser{}, &testProduct{}})
					return cfg
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Open(tt.args.dsn, tt.args.models...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Open() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpen_InvalidModel(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Open() did not panic")
		}
	}()

	Open(dsn, &invalidPublicUser{}, &invalidPrivateProduct{}, &invalidNonTabler{})
}

func TestNew(t *testing.T) {
	config := Config{
		DSN: dsn,
	}

	models := []interface{}{&testUser{}, &testProduct{}}

	want := &Dialector{
		rw:        &sync.RWMutex{},
		Dialector: *postgres.New(config).(*postgres.Dialector),
		migratorConfig: func() *migratorConfig {
			cfg, _ := newMigratorConfig(models)
			return cfg
		}(),
	}

	got := New(config, models...)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("New() = %v, want %v", got, want)
	}
}

func TestNew_InvalidModel(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("New() did not panic")
		}
	}()

	config := Config{
		DSN: dsn,
	}

	models := []interface{}{&invalidPublicUser{}, &invalidPrivateProduct{}, &invalidNonTabler{}}

	New(config, models...)
}

func TestNewMultitenancyConfig(t *testing.T) {
	models := []interface{}{&testUser{}, &testProduct{}}

	wantPublicModels := []interface{}{&testUser{}}
	wantPrivateModels := []interface{}{&testProduct{}}

	// got := newMultitenancyConfig(models)
	got, err := newMigratorConfig(models)
	if err != nil {
		t.Errorf("newMultitenancyConfig() error = %v", err)
	}

	if !reflect.DeepEqual(got.publicModels, wantPublicModels) {
		t.Errorf("newMultitenancyConfig() got publicModels = %v, want %v", got.publicModels, wantPublicModels)
	}

	if !reflect.DeepEqual(got.tenantModels, wantPrivateModels) {
		t.Errorf("newMultitenancyConfig() got tenantModels = %v, want %v", got.tenantModels, wantPrivateModels)
	}
}

func TestNewMultitenancyConfig_InvalidModel(t *testing.T) {
	models := []interface{}{&invalidPublicUser{}, &invalidPrivateProduct{}, &invalidNonTabler{}}

	for _, model := range models {
		_, err := newMigratorConfig([]interface{}{model})
		if err == nil {
			t.Errorf("newMultitenancyConfig() error = %v, wantErr %v", err, true)
		}
	}
}

func TestDialector_Migrator(t *testing.T) {
	type fields struct {
		Dialector          postgres.Dialector
		multitenancyConfig *migratorConfig
	}
	type args struct {
		db *gorm.DB
	}
	db := testutil.NewTestDB()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   gorm.Migrator
	}{
		{
			name: "Test Migrator",
			fields: fields{
				Dialector: *db.Dialector.(*postgres.Dialector),
				multitenancyConfig: &migratorConfig{
					publicModels: []interface{}{&testUser{}},
					tenantModels: []interface{}{&testProduct{}},
				},
			},
			args: args{
				db: db,
			},
			want: &Migrator{
				rw: &sync.RWMutex{},
				Migrator: postgres.Migrator{
					Migrator: migrator.Migrator{
						Config: migrator.Config{
							DB:                          db,
							Dialector:                   *db.Dialector.(*postgres.Dialector),
							CreateIndexAfterCreateTable: true,
						},
					},
				},
				migratorConfig: &migratorConfig{
					publicModels: []interface{}{&testUser{}},
					tenantModels: []interface{}{&testProduct{}},
				},
			},
		},
	}

	compareMigrators := func(x, y gorm.Migrator) bool {
		return reflect.DeepEqual(x.(*Migrator).migratorConfig, y.(*Migrator).migratorConfig)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialector := Dialector{
				Dialector:      tt.fields.Dialector,
				migratorConfig: tt.fields.multitenancyConfig,
				rw:             &sync.RWMutex{},
			}
			got := dialector.Migrator(tt.args.db)
			if !compareMigrators(got, tt.want) {
				t.Errorf("Dialector.Migrator() mismatch, want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func TestRegisterModels(t *testing.T) {
	db, err := gorm.Open(Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	type args struct {
		db     *gorm.DB
		models []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test RegisterModels",
			args: args{
				db:     db,
				models: []interface{}{&testUser{}, &testProduct{}},
			},
		},
		{
			name: "Test RegisterModels with invalid model",
			args: args{
				db:     db,
				models: []interface{}{&invalidPublicUser{}, &invalidPrivateProduct{}, &invalidNonTabler{}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterModels(tt.args.db, tt.args.models...)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			dialector := db.Dialector.(*Dialector)
			if !reflect.DeepEqual(dialector.migratorConfig.publicModels, []interface{}{&testUser{}}) {
				t.Errorf("RegisterModels() publicModels = %v, want %v", dialector.migratorConfig.publicModels, []interface{}{&testUser{}})
			}
			if !reflect.DeepEqual(dialector.migratorConfig.tenantModels, []interface{}{&testProduct{}}) {
				t.Errorf("RegisterModels() tenantModels = %v, want %v", dialector.migratorConfig.tenantModels, []interface{}{&testProduct{}})
			}
		})
	}
}

func TestCreateSchemaForTenant(t *testing.T) {
	// Create a test schema name
	schemaName := "test_schema"

	// Create a new GORM DB instance
	db, err := gorm.Open(Open(dsn, &testUser{}, &testProduct{}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create the schema for the tenant
	err = CreateSchemaForTenant(db, schemaName)
	if err != nil {
		t.Fatalf("Failed to create schema for tenant: %v", err)
	}
	t.Cleanup(func() {
		// Drop the schema if exists
		err = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error
		if err != nil {
			t.Fatalf("Failed to drop schema for tenant: %v", err)
		}
	})

	// Check if the schema exists
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = ?)", schemaName).Scan(&exists).Error
	if err != nil {
		t.Fatalf("Failed to check if schema exists: %v", err)
	}
}

func TestMigratePublicSchema(t *testing.T) {
	// Create a new GORM DB instance
	db, err := gorm.Open(Open(dsn, &testUser{}, &testProduct{}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate the public schema
	err = MigratePublicSchema(db)
	if err != nil {
		t.Fatalf("Failed to migrate public schema: %v", err)
	}
	t.Cleanup(func() {
		// Drop public tables
		err = db.Migrator().DropTable(&testUser{}, &testProduct{})
		if err != nil {
			t.Fatalf("Failed to drop public tables: %v", err)
		}
	})
}

func TestDropSchemaForTenant(t *testing.T) {
	// Create a test schema name
	schemaName := "test_schema"

	// Create a new GORM DB instance
	db, err := gorm.Open(Open(dsn, &testUser{}, &testProduct{}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create the schema for the tenant
	err = CreateSchemaForTenant(db, schemaName)
	if err != nil {
		t.Fatalf("Failed to create schema for tenant: %v", err)
	}

	// Drop the schema for the tenant
	err = DropSchemaForTenant(db, schemaName)
	if err != nil {
		t.Fatalf("Failed to drop schema for tenant: %v", err)
	}

	// Check if the schema exists
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = ?)", schemaName).Scan(&exists).Error
	if err != nil {
		t.Fatalf("Failed to check if schema exists: %v", err)
	}
	if exists {
		t.Fatalf("Failed to drop schema for tenant: schema still exists")
	}
}
