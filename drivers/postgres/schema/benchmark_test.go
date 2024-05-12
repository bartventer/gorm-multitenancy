package schema_test

import (
	"testing"

	pgdriver "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
	pgschema "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema"
	"github.com/bartventer/gorm-multitenancy/v6/internal/testutil"
	scopes "github.com/bartventer/gorm-multitenancy/v6/scopes"
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

func BenchmarkScopingQueries(b *testing.B) {
	db, tenant, cleanup, err := setupScopingBenchmark(b, testutil.WithDBName("tenants1"))
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}
	b.Cleanup(cleanup)
	b.ResetTimer()
	tests := []struct {
		name             string
		preRun           func(*testing.B, *gorm.DB, *Tenant)
		setup            func(*testing.B, *gorm.DB, *Tenant)
		withTenantSchema func(*testing.B, *gorm.DB, *Tenant)
		setSearchPath    func(*testing.B, *gorm.DB, *Tenant)
	}{
		{
			name: "Create",
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
			},
		},
		{
			name:   "Find",
			preRun: createBook,
			withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
				var book Book
				if err := d.Scopes(scopes.WithTenantSchema(t.SchemaName)).First(&book).Error; err != nil {
					b.Errorf("Find book failed, expected %v, got %v", nil, err)
				}
			},
			setSearchPath: func(b *testing.B, d *gorm.DB, t *Tenant) {
				_, resetSearchPath := pgschema.SetSearchPath(d, t.SchemaName)
				if err := d.Error; err != nil {
					b.Fatalf("SetSearchPath() with valid schema name failed: %v", err)
				}
				defer resetSearchPath()
				var book Book
				if err := d.First(&book).Error; err != nil {
					b.Errorf("Find book failed, expected %v, got %v", nil, err)
				}
			},
		},
		{
			name:   "Update",
			preRun: createBook,
			withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
				var book Book
				if err := d.Scopes(scopes.WithTenantSchema(t.SchemaName)).First(&book).Error; err != nil {
					b.Fatalf("Find book failed: %v", err)
				}
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
				var book Book
				if err := d.First(&book).Error; err != nil {
					b.Fatalf("Find book failed: %v", err)
				}
			},
		},
		{
			name:  "Delete",
			setup: createBook,
			withTenantSchema: func(b *testing.B, d *gorm.DB, t *Tenant) {
				var book Book
				if err := d.Scopes(scopes.WithTenantSchema(t.SchemaName)).First(&book).Error; err != nil {
					b.Fatalf("Find book failed: %v", err)
				}
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
				var book Book
				if err := d.First(&book).Error; err != nil {
					b.Fatalf("Find book failed: %v", err)
				}
				if err := d.Delete(&book).Error; err != nil {
					b.Errorf("Delete book failed, expected %v, got %v", nil, err)
				}
			},
		},
	}
	for _, tt := range tests {
		if tt.preRun != nil {
			b.StopTimer()
			tt.preRun(b, db, tenant)
			b.StartTimer()
		}
		b.Run(tt.name+"/WithTenantSchema", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if tt.setup != nil {
					b.StopTimer()
					tt.setup(b, db, tenant)
					b.StartTimer()
				}
				tt.withTenantSchema(b, db, tenant)
			}
		})
		b.Run(tt.name+"/SetSearchPath", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if tt.setup != nil {
					b.StopTimer()
					tt.setup(b, db, tenant)
					b.StartTimer()
				}
				tt.setSearchPath(b, db, tenant)
			}
		})
	}
}
