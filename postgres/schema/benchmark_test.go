package schema_test

import (
	"testing"

	pgdriver "github.com/bartventer/gorm-multitenancy/postgres/v7"
	"github.com/bartventer/gorm-multitenancy/postgres/v7/internal/testutil"
	pgschema "github.com/bartventer/gorm-multitenancy/postgres/v7/schema"
	scopes "github.com/bartventer/gorm-multitenancy/postgres/v7/scopes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Tenant struct {
	pgdriver.TenantPKModel
}

func (Tenant) TableName() string { return "public.tenants" }

type Book struct {
	gorm.Model
	Title    string
	TenantID string `gorm:"column:tenant_schema"`
	Tenant   Tenant `gorm:"foreignKey:TenantID;references:SchemaName"`
}

func (Book) TableName() string   { return "books" }
func (Book) IsTenantTable() bool { return true }

func BenchmarkScopingQueries(b *testing.B) {
	db, tenant, cleanup, err := setupScopingBenchmark(b, testutil.WithDBName("tenants1"))
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	b.Cleanup(cleanup)
	type args struct {
		preRun, setup, withTenantSchema, setSearchPath RunFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create",
			args: args{
				withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
					book := Book{
						Tenant: *tenant,
					}
					if err := db.Scopes(scopes.WithTenantSchema(tenant.SchemaName)).Create(&book).Error; err != nil {
						b.Errorf("Create book failed, expected %v, got %v", nil, err)
					}
				},
				setSearchPath: func(b *testing.B, d *gorm.DB, t *Tenant) {
					_, resetSearchPath := pgschema.SetSearchPath(db, tenant.SchemaName)
					if err := db.Error; err != nil {
						b.Fatalf("SetSearchPath() with valid schema name failed: %v", err)
					}
					defer resetSearchPath()
					book := Book{
						Tenant: *tenant,
					}
					if err := db.Create(&book).Error; err != nil {
						b.Errorf("Create book failed, expected %v, got %v", nil, err)
					}
				},
			},
		},
		{
			name: "Find",
			args: args{
				preRun: createBook,
				withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
					_ = findBook(b, d.Scopes(scopes.WithTenantSchema(t.SchemaName)))
				},
				setSearchPath: func(b *testing.B, d *gorm.DB, t *Tenant) {
					_, resetSearchPath := pgschema.SetSearchPath(d, t.SchemaName)
					if err := d.Error; err != nil {
						b.Fatalf("SetSearchPath() with valid schema name failed: %v", err)
					}
					defer resetSearchPath()
					_ = findBook(b, d)
				},
			},
		},
		{
			name: "Update",
			args: args{
				preRun: createBook,
				withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
					book := findBook(b, d.Scopes(scopes.WithTenantSchema(t.SchemaName)))
					book.Title = "Updated"
					if err := d.Scopes(scopes.WithTenantSchema(t.SchemaName)).Save(&book).Error; err != nil {
						b.Errorf("Update book failed, expected %v, got %v", nil, err)
					}
				},
				setSearchPath: func(b *testing.B, d *gorm.DB, t *Tenant) {
					_, resetSearchPath := pgschema.SetSearchPath(d, t.SchemaName)
					if err := d.Error; err != nil {
						b.Fatalf("SetSearchPath() with valid schema name failed: %v", err)
					}
					defer resetSearchPath()
					book := findBook(b, d)
					book.Title = "Updated"
					if err := d.Save(&book).Error; err != nil {
						b.Errorf("Update book failed, expected %v, got %v", nil, err)
					}
				},
			},
		},
		{
			name: "Delete",
			args: args{
				setup: createBook,
				withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
					book := findBook(b, d.Scopes(scopes.WithTenantSchema(t.SchemaName)))
					if err := d.Scopes(scopes.WithTenantSchema(t.SchemaName)).Delete(&book).Error; err != nil {
						b.Errorf("Delete book failed, expected %v, got %v", nil, err)
					}
				},
				setSearchPath: func(b *testing.B, d *gorm.DB, t *Tenant) {
					_, resetSearchPath := pgschema.SetSearchPath(d, t.SchemaName)
					if err := d.Error; err != nil {
						b.Fatalf("SetSearchPath() with valid schema name failed: %v", err)
					}
					defer resetSearchPath()
					book := findBook(b, d)
					if err := d.Delete(&book).Error; err != nil {
						b.Errorf("Delete book failed, expected %v, got %v", nil, err)
					}
				},
			},
		},
	}
	b.ResetTimer()
	for _, tt := range tests {
		if tt.args.preRun != nil {
			b.StopTimer()
			tt.args.preRun(b, db, tenant)
			b.StartTimer()
		}
		runScopingBenchmark(b, tt.name+"/SetSearchPath", db, tenant, tt.args.setup, tt.args.setSearchPath)
		runScopingBenchmark(b, tt.name+"/WithTenantSchema", db, tenant, tt.args.setup, tt.args.withTenantSchema)
	}
}

// RunFunc is a function that runs a benchmark test.
type RunFunc func(*testing.B, *gorm.DB, *Tenant)

func setupScopingBenchmark(b *testing.B, dsnOpt testutil.DSNOption) (*gorm.DB, *Tenant, func(), error) {
	b.Helper()
	db, err := gorm.Open(pgdriver.Open(testutil.GetDSN(dsnOpt)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := pgdriver.RegisterModels(db, &Tenant{}, &Book{}); err != nil {
		return nil, nil, nil, err
	}
	if err := pgdriver.MigratePublicSchema(db); err != nil {
		return nil, nil, nil, err
	}
	tenant := Tenant{
		TenantPKModel: pgdriver.TenantPKModel{
			DomainURL:  "tenant1.example.com",
			SchemaName: "tenant1",
		},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tenant).Error; err != nil {
		return nil, nil, nil, err
	}
	if err := pgdriver.CreateSchemaForTenant(db, tenant.SchemaName); err != nil {
		return nil, nil, nil, err
	}
	cleanup := func() {
		if err := pgdriver.DropSchemaForTenant(db, tenant.SchemaName); err != nil {
			b.Fatalf("Drop schema for tenant failed: %v", err)
		}
	}
	return db, &tenant, cleanup, nil
}

func createBook(b *testing.B, db *gorm.DB, tenant *Tenant) {
	b.Helper()
	book := Book{
		Tenant: *tenant,
	}
	if err := db.Scopes(scopes.WithTenantSchema(tenant.SchemaName)).Create(&book).Error; err != nil {
		b.Errorf("Create book failed, expected %v, got %v", nil, err)
	}
}

// findBook finds a book.
func findBook(b *testing.B, db *gorm.DB) Book {
	b.Helper()
	var book Book
	if err := db.First(&book).Error; err != nil {
		b.Fatalf("Find book failed: %v", err)
	}
	return book
}

// runScopingBenchmark runs a benchmark test.
func runScopingBenchmark(b *testing.B, name string, db *gorm.DB, tenant *Tenant, setup, fn RunFunc) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if setup != nil {
				b.StopTimer()
				setup(b, db, tenant)
				b.StartTimer()
			}
			fn(b, db, tenant)
		}
	})
}
