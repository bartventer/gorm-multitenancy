package postgres

import (
	"fmt"
	"strings"
	"sync"

	multitenancy "github.com/bartventer/gorm-multitenancy/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

type (
	// Config is the configuration for the postgres driver
	Config = postgres.Config

	// Dialector is the postgres dialector with multitenancy support
	Dialector struct {
		postgres.Dialector
		*multitenancyConfig
		rw *sync.RWMutex
	}
)

// Check interface
var _ gorm.Dialector = (*Dialector)(nil)

// Open creates a new postgres dialector with multitenancy support
func Open(dsn string, models ...interface{}) gorm.Dialector {
	d := &Dialector{
		Dialector: *postgres.Open(dsn).(*postgres.Dialector),
		rw:        &sync.RWMutex{},
	}
	mtc, err := newMultitenancyConfig(models)
	if err != nil {
		panic(err)
	}
	d.multitenancyConfig = mtc
	return d
}

// New creates a new postgres dialector with multitenancy support
func New(config Config, models ...interface{}) gorm.Dialector {
	d := &Dialector{
		Dialector: *postgres.New(config).(*postgres.Dialector),
		rw:        &sync.RWMutex{},
	}
	mtc, err := newMultitenancyConfig(models)
	if err != nil {
		panic(err)
	}
	d.multitenancyConfig = mtc
	return d
}

// Migrator returns the migrator with multitenancy support
func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	dialector.rw.RLock()
	defer dialector.rw.RUnlock()

	return &Migrator{
		postgres.Migrator{
			Migrator: migrator.Migrator{
				Config: migrator.Config{
					DB:                          db,
					Dialector:                   dialector,
					CreateIndexAfterCreateTable: true,
				},
			},
		},
		&multitenancyConfig{
			publicModels: dialector.publicModels,
			tenantModels: dialector.tenantModels,
			models:       dialector.models,
		},
		&sync.RWMutex{},
	}
}

// RegisterModels registers the models for multitenancy
func RegisterModels(db *gorm.DB, models ...interface{}) error {
	dialector := db.Dialector.(*Dialector)
	mtc, err := newMultitenancyConfig(models)
	if err != nil {
		return fmt.Errorf("failed to register models: %w", err)
	}

	dialector.rw.Lock()
	dialector.multitenancyConfig = mtc
	dialector.rw.Unlock()
	return nil
}

// MigratePublicSchema migrates the public tables
func MigratePublicSchema(db *gorm.DB) error {
	return db.Migrator().(*Migrator).MigratePublicSchema()
}

// CreateSchemaForTenant creates the schema for the tenant, and migrates the private tables
func CreateSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).CreateSchemaForTenant(schemaName)
}

// DropSchemaForTenant drops the schema for the tenant (CASCADING tables)
func DropSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).DropSchemaForTenant(schemaName)
}

// newMultitenancyConfig creates a new multitenancy configuration based on the provided models.
// It separates the models into public models and private models based on their table names.
// Public models are those that have a table name starting with the default schema name (PublicSchemaName),
// while private models are those that implement the TenantTabler interface and have a table name without a fullstop.
// If any model does not meet the required criteria, an error message is appended to the errStrings slice.
// If there are any errors, the function returns nil and an error. Otherwise, it returns a new multitenancyConfig
// containing the public models, private models, and all the models.
func newMultitenancyConfig(models []interface{}) (*multitenancyConfig, error) {
	var (
		publicModels  = make([]interface{}, 0, len(models))
		privateModels = make([]interface{}, 0, len(models))
		errStrings    = make([]string, 0)
	)
	for _, m := range models {
		tn, ok := m.(interface{ TableName() string })
		if !ok {
			errStrings = append(errStrings, fmt.Sprintf("model %T does not implement TableName()", m))
			continue
		}
		tt, ok := m.(multitenancy.TenantTabler)
		parts := strings.Split(tn.TableName(), ".")
		if ok && tt.IsTenantTable() {
			// ensure that the private model does not contain a fullstop
			if len(parts) > 1 {
				errStrings = append(errStrings, fmt.Sprintf("invalid table name for model %T labeled as tenant table, table name should not contain a fullstop, got '%s'", m, tn.TableName()))
				continue
			}
			privateModels = append(privateModels, m)
		} else {
			// ensure that the public model starts with the default schema (public.)
			if len(parts) != 2 || parts[0] != PublicSchemaName {
				errStrings = append(errStrings, fmt.Sprintf("invalid table name for model %T labeled as public table, table name should start with '%s.', got '%s'", m, PublicSchemaName, tn.TableName()))
				continue
			}
			publicModels = append(publicModels, m)
		}
	}

	// if there are errors, panic
	if len(errStrings) > 0 {
		return nil, fmt.Errorf("failed to create multitenancy config: %s", strings.Join(errStrings, "; "))
	}

	return &multitenancyConfig{
		publicModels: publicModels,
		tenantModels: privateModels,
		models:       models,
	}, nil
}
