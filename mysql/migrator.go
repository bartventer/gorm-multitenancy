package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bartventer/gorm-multitenancy/mysql/v7/schema"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/gmterrors"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

var _ gorm.Migrator = new(Migrator)

func inTransaction(db *gorm.DB) bool {
	committer, ok := db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}

func (m Migrator) AutoMigrate(values ...interface{}) error {
	// Check if the migration option is set
	_, err := migrationOptionFromDB(m.DB)
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, err)
	}
	return m.Migrator.AutoMigrate(values...)
}

// MigrateTenantModels creates a database for a specific tenant and migrates the tenant tables.
func (m Migrator) MigrateTenantModels(tenantID string) error {
	if inTransaction(m.DB) {
		return m.migrateTenantModels(tenantID)
	} else {
		return m.DB.Transaction(func(tx *gorm.DB) error {
			return m.migrateTenantModels(tenantID)
		})
	}
}

func (m Migrator) migrateTenantModels(tenantID string) error {
	m.logger.Printf("⏳ migrating tables for tenant %s", tenantID)
	tenantModels := m.registry.TenantModels
	if len(tenantModels) == 0 {
		return gmterrors.NewWithScheme(DriverName, errors.New("no tenant tables to migrate"))
	}
	tx := m.DB.Session(&gorm.Session{})
	err := tx.Exec("CREATE DATABASE IF NOT EXISTS " + tx.Statement.Quote(tenantID)).Error
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to create database for tenant %s: %w", tenantID, err))
	}

	// Use tenant's database
	var reset func() error
	reset, err = schema.UseDatabase(tx, tenantID)
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to switch to tenant database %s: %w", tenantID, err))
	}
	defer reset()

	// Migrate tenant tables
	if err = tx.Scopes(withMigrationOption(migratorOption)).
		AutoMigrate(driver.ModelsToInterfaces(tenantModels)...); err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to migrate tables for tenant %s: %w", tenantID, err))
	}
	m.logger.Printf("✅ private tables migrated for tenant %s", tenantID)
	return nil
}

// MigrateSharedModels migrates the shared tables in the database.
func (m Migrator) MigrateSharedModels() error {
	m.logger.Println("⏳ migrating public tables")
	publicModels := m.registry.SharedModels
	if len(publicModels) == 0 {
		return gmterrors.NewWithScheme(DriverName, errors.New("no public tables to migrate"))
	}

	// Create public table if it doesn't exist
	if err := m.DB.Exec("CREATE DATABASE IF NOT EXISTS public").Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to create public database: %w", err))
	}

	// switch to public table
	if err := m.DB.Exec("USE public").Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to switch to public database: %w", err))
	}

	var err error
	if err = m.DB.Scopes(withMigrationOption(migratorOption)).
		AutoMigrate(driver.ModelsToInterfaces(publicModels)...); err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to migrate public tables: %w", err))
	}
	m.logger.Println("✅ public tables migrated")
	return nil
}

// DropDatabaseForTenant drops the database for a specific tenant.
func (m Migrator) DropDatabaseForTenant(tenant string) error {
	if inTransaction(m.DB) {
		return m.dropDatabaseForTenant(tenant)
	} else {
		return m.DB.Transaction(func(tx *gorm.DB) error {
			return m.dropDatabaseForTenant(tenant)
		})
	}
}

func (m Migrator) dropDatabaseForTenant(tenant string) error {
	tx := m.DB.Session(&gorm.Session{})
	m.logger.Printf("⏳ dropping database for tenant %s", tenant)
	var err error
	if err = tx.Exec("DROP DATABASE IF EXISTS " + tx.Statement.Quote(tenant)).Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to drop database for tenant %s: %w", tenant, err))
	}
	m.logger.Printf("✅ database dropped for tenant %s", tenant)
	return nil
}

// Note: Subject to removal if the below changes are integrated into the GORM MySQL driver.
func (m Migrator) queryRaw(sql string, values ...interface{}) (tx *gorm.DB) {
	queryTx := m.DB
	if m.DB.DryRun {
		queryTx = m.DB.Session(&gorm.Session{})
		queryTx.DryRun = false
	}
	return queryTx.Raw(sql, values...)
}

// Potential enhancement for GORM. Aims to simplify database name retrieval.
//
// Note: May be removed if integrated into GORM's MySQL driver.
//
// TODO: Create PR for these changes and link to issue https://github.com/go-gorm/gorm/issues/3958.
func (m Migrator) CurrentDatabase() (name string) {
	m.DB.Raw("SELECT DATABASE()").Row().Scan(&name)
	return
}

// Enhances schema and table identification for MySQL.
//
// Note: Subject to removal if adopted into the GORM MySQL driver.
//
// TODO: Create PR for these changes, addressing https://github.com/go-gorm/gorm/issues/3958.
func (m Migrator) CurrentSchema(stmt *gorm.Statement, table string) (interface{}, interface{}) {
	if strings.Contains(table, ".") {
		if tables := strings.Split(table, `.`); len(tables) == 2 {
			return tables[0], tables[1]
		}
	}

	if stmt.TableExpr != nil {
		if tables := strings.Split(stmt.TableExpr.SQL, "`.`"); len(tables) == 2 {
			return strings.TrimPrefix(tables[0], "`"), table
		}
	}
	return m.CurrentDatabase(), table
}

// Improves table existence check for MySQL, enhancing efficiency and accuracy.
//
// Note: Considered for deprecation if merged into GORM's MySQL driver.
//
// TODO: Create PR for this change, addressing issue https://github.com/go-gorm/gorm/issues/3958.
func (m Migrator) HasTable(value interface{}) bool {
	var exists bool
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		existSQL := `
            SELECT EXISTS (
                SELECT 1
                FROM INFORMATION_SCHEMA.tables 
                WHERE table_schema = ? AND table_name = ? AND table_type = 'BASE TABLE'
            )
        `
		return m.queryRaw(existSQL, currentSchema, curTable).Row().Scan(&exists)
	})
	return exists
}

// Adds MySQL-specific index existence check.
//
// Note: May be deprecated if integrated into GORM's MySQL driver.
//
// TODO: Raise issue for discussion.
func (m Migrator) HasIndex(value interface{}, name string) bool {
	var exists bool
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}
		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		existSQL := `
            SELECT EXISTS (
                SELECT 1
                FROM INFORMATION_SCHEMA.statistics
                WHERE table_name = ? AND index_name = ? AND table_schema = ?
            )
        `
		return m.queryRaw(existSQL, curTable, name, currentSchema).Row().Scan(&exists)
	})
	return exists
}

// Implements MySQL-specific constraint existence check.
// Address `Error 3822 (HY000): Duplicate check constraint name '...'` issue when
// creating a table with a check constraint in MySQL databases.
//
// Note: Potential for deprecation if merged into GORM's MySQL driver.
//
// TODO: Raise issue for discussion.
func (m Migrator) HasConstraint(value interface{}, name string) bool {
	var exists bool
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, table := m.GuessConstraintInterfaceAndTable(stmt, name)
		if constraint != nil {
			name = constraint.GetName()
		}
		currentSchema, curTable := m.CurrentSchema(stmt, table)
		existSQL := `
            SELECT EXISTS (
                SELECT 1
                FROM INFORMATION_SCHEMA.table_constraints
                WHERE table_schema = ? AND table_name = ? AND constraint_name = ?
            )
        `
		return m.queryRaw(existSQL, currentSchema, curTable, name).Row().Scan(&exists)
	})

	return exists
}

// Adds MySQL-specific column existence check.
//
// Note: May be removed if adopted into GORM's MySQL driver.
//
// TODO: Raise issue for discussion.
func (m Migrator) HasColumn(value interface{}, field string) bool {
	var exists bool
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		name := field
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				name = field.DBName
			}
		}

		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		existSQL := `
            SELECT EXISTS (
                SELECT 1
                FROM INFORMATION_SCHEMA.columns
                WHERE table_schema = ? AND table_name = ? AND column_name = ?
            )
        `
		return m.queryRaw(existSQL, currentSchema, curTable, name).Row().Scan(&exists)
	})

	return exists
}

// Refines ColumnTypes method for MySQL, considering specific data type details.
//
// Note: Subject to deprecation if integrated into GORM's MySQL driver.
//
// TODO: Raise issue for discussion.
func (m Migrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columnTypes := make([]gorm.ColumnType, 0)
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var (
			currentDatabase, table = m.CurrentSchema(stmt, stmt.Table)
			columnTypeSQL          = "SELECT column_name, column_default, is_nullable = 'YES', data_type, character_maximum_length, column_type, column_key, extra, column_comment, numeric_precision, numeric_scale "
			rows, err              = m.DB.Session(&gorm.Session{}).Table(stmt.Quote(currentDatabase) + "." + stmt.Quote(table)).Limit(1).Rows()
		)

		if err != nil {
			return err
		}

		rawColumnTypes, err := rows.ColumnTypes()

		if err != nil {
			return err
		}

		if err := rows.Close(); err != nil {
			return err
		}
		if !m.Migrator.DisableDatetimePrecision {
			columnTypeSQL += ", datetime_precision "
		}
		columnTypeSQL += "FROM INFORMATION_SCHEMA.columns WHERE table_schema = ? AND table_name = ? ORDER BY ORDINAL_POSITION"

		columns, rowErr := m.DB.Table(table.(string)).Raw(columnTypeSQL, currentDatabase, table).Rows()
		if rowErr != nil {
			return rowErr
		}

		defer columns.Close()

		for columns.Next() {
			var (
				column            migrator.ColumnType
				datetimePrecision sql.NullInt64
				extraValue        sql.NullString
				columnKey         sql.NullString
				values            = []interface{}{
					&column.NameValue, &column.DefaultValueValue, &column.NullableValue, &column.DataTypeValue, &column.LengthValue, &column.ColumnTypeValue, &columnKey, &extraValue, &column.CommentValue, &column.DecimalSizeValue, &column.ScaleValue,
				}
			)

			if !m.Migrator.DisableDatetimePrecision {
				values = append(values, &datetimePrecision)
			}

			if scanErr := columns.Scan(values...); scanErr != nil {
				return scanErr
			}

			column.PrimaryKeyValue = sql.NullBool{Bool: false, Valid: true}
			column.UniqueValue = sql.NullBool{Bool: false, Valid: true}
			switch columnKey.String {
			case "PRI":
				column.PrimaryKeyValue = sql.NullBool{Bool: true, Valid: true}
			case "UNI":
				column.UniqueValue = sql.NullBool{Bool: true, Valid: true}
			}

			if strings.Contains(extraValue.String, "auto_increment") {
				column.AutoIncrementValue = sql.NullBool{Bool: true, Valid: true}
			}

			// only trim paired single-quotes
			s := column.DefaultValueValue.String
			for (len(s) >= 3 && s[0] == '\'' && s[len(s)-1] == '\'' && s[len(s)-2] != '\\') ||
				(len(s) == 2 && s == "''") {
				s = s[1 : len(s)-1]
			}
			column.DefaultValueValue.String = s
			if m.Dialector.DontSupportNullAsDefaultValue {
				// rewrite mariadb default value like other version
				if column.DefaultValueValue.Valid && column.DefaultValueValue.String == "NULL" {
					column.DefaultValueValue.Valid = false
					column.DefaultValueValue.String = ""
				}
			}

			if datetimePrecision.Valid {
				column.DecimalSizeValue = datetimePrecision
			}

			for _, c := range rawColumnTypes {
				if c.Name() == column.NameValue.String {
					column.SQLColumnType = c
					break
				}
			}

			columnTypes = append(columnTypes, column)
		}

		return nil
	})

	return columnTypes, err
}
