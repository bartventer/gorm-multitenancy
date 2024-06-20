/*
Package scopes provides a set of predefined multitenancy scopes for GORM.
*/
package scopes

import (
	"errors"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// WithTenantSchema alters the table name to prefix it with the tenant schema.
//
// The table name is retrieved from the statement if set manually, otherwise
// attempts to get it from the model or destination.
//
// Example with table name set manually:
//
//	type Book struct { ... } // does not implement Tabler interface, must set TableName manually
//
//	db.Table("books").Scopes(WithTenantSchema("tenant2")).Find(&Book{})
//	// SELECT * FROM "tenant2"."books"
//
// Example with Tabler interface:
//
//	type Book struct { ... }
//
//	func (u *Book) TableName() string { return "books" } // implements Tabler interface, no need to set TableName manually
//
//	db.Scopes(WithTenantSchema("tenant1")).Find(&Book{})
//	// SELECT * FROM "tenant1"."books"
//
// Example with model set manually:
//
//	type Book struct { ... }
//
//	db.Model(&Book{}).Scopes(WithTenantSchema("tenant1")).Find(&Book{}) // model is set manually.
//	// SELECT * FROM "tenant1"."books"
//
// Example with destination set to a pointer to a struct:
//
//	type Book struct { ... }
//
//	db.Scopes(WithTenantSchema("tenant1")).Find(&Book{}) // destination is set to a pointer to a struct.
//	// SELECT * FROM "tenant1"."books"
//
// Example with destination set to a pointer to an array/slice:
//
//	type Book struct { ... }
//
//	db.Scopes(WithTenantSchema("tenant1")).Find(&[]Book{}) // destination is set to a pointer to an array/slice.
//	// SELECT * FROM "tenant1"."books"
func WithTenantSchema(tenant string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var tableName string
		switch {
		case db.Statement.Table != "":
			tableName = db.Statement.Table
		case db.Statement.Model != nil:
			tableName = tableNameFromInterface(db.Statement.Model)
		case db.Statement.Dest != nil:
			var err error
			tableName, err = tableNameFromReflectValue(reflect.ValueOf(db.Statement.Dest))
			if err != nil {
				_ = db.AddError(err)
				return db
			}
		}
		if tableName != "" {
			return db.Table(tenant + "." + tableName)
		}
		_ = db.AddError(gorm.ErrModelValueRequired)
		return db
	}
}

// tableNameFromInterface returns the table name from an interface.
func tableNameFromInterface(val interface{}) string {
	if s, ok := val.(schema.Tabler); ok {
		return s.TableName()
	}
	return ""
}

// tableNameFromReflectValue returns the table name from a [reflect.Value].
func tableNameFromReflectValue(valPtr reflect.Value) (string, error) {
	if valPtr.Kind() != reflect.Ptr {
		return "", errors.New("destination must be a pointer")
	}
	val := valPtr.Elem()
	switch val.Kind() { //nolint:exhaustive // only interested in a struct and a array/slice
	case reflect.Struct:
		return tableNameFromInterface(val.Interface()), nil
	case reflect.Slice, reflect.Array:
		return tableNameFromInterface(reflect.New(val.Type().Elem()).Interface()), nil
	default:
		return "", nil
	}
}
