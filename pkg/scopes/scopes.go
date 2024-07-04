/*
Package scopes provides a set of predefined multitenancy scopes for GORM.
*/
package scopes

import (
	"reflect"

	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"gorm.io/gorm"
)

var tenantTablerType = reflect.TypeFor[driver.TenantTabler]()

// WithTenantSchema returns a GORM scope function that prefixes the table name with
// the specified tenant schema.
//
// The function supports different strategies for determining the table name, which
// are attempted in the following order:
//  1. Direct specification via [gorm.DB.Table].
//  2. Model type set via [gorm.DB.Model] implementing [driver.TenantTabler].
//  3. Destination type implementing [driver.TenantTabler].
//
// If the table name cannot be determined, an error is added to the DB instance.
//
// Examples (assuming the following model):
//
//	type Book struct { ... }
//	var _ driver.TenantTabler = new(Book)
//	func (Book) TableName() string { return "books" }
//	func (Book) IsSharedModel() bool { ... }
//
// Direct table specification:
//
//	db.Table("books").Scopes(WithTenantSchema("tenant1")).Find(...)
//	// SELECT * FROM "tenant1"."books"
//
// Inferred from `Model` call:
//
//	db.Model(&Book{}).Scopes(WithTenantSchema("tenant1")).Find(...)
//	// SELECT * FROM "tenant1"."books"
//
// Inferred from the destination type:
//
//	db.Scopes(WithTenantSchema("tenant1")).Find(&[]Book{})
//	// SELECT * FROM "tenant1"."books"
func WithTenantSchema(tenant string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		stmt := db.Statement
		if stmt.Table != "" {
			return db.Table(tenant + "." + stmt.Table)
		}

		var tableName string
		switch {
		case db.Statement.Model != nil:
			tableName = tableNameFromInterface(db.Statement.Model)
		case db.Statement.Dest != nil:
			tableName = tableNameFromInterface(db.Statement.Dest)
		}
		if tableName != "" {
			return db.Table(tenant + "." + tableName)
		}
		_ = db.AddError(gorm.ErrModelValueRequired)
		return db
	}
}

// tableNameFromInterface attempts to determine the table name from the provided interface value.
// It supports values that directly implement the `driver.TenantTabler` interface or are struct types
// that could potentially implement the interface. For slices or arrays, it attempts to infer the table name
// from the element type.
func tableNameFromInterface(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return ""
	case driver.TenantTabler:
		return v.TableName()
	}

	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	switch kind := rv.Kind(); kind { //nolint:exhaustive // Only interested in a subset of kinds
	case reflect.Struct:
		return "" // Already handled by the driver.TenantTabler case
	case reflect.Slice, reflect.Array:
		if rv.Len() > 0 {
			rv = rv.Index(0)
		} else {
			elemType := rv.Type().Elem()
			if elemType.Kind() == reflect.Ptr {
				rv = reflect.New(elemType.Elem())
			} else {
				rv = reflect.New(elemType).Elem()
			}
		}
	}

	if rv.Type().Implements(tenantTablerType) && rv.CanInterface() {
		return rv.Interface().(driver.TenantTabler).TableName()
	}

	return ""
}
