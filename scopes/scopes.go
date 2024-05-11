/*
Package scopes provides a set of predefined GORM scopes for managing multi-tenant applications using the gorm-multitenancy library.
*/
package scopes

import (
	"reflect"
	"sync"

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
//	db.Table("books").Scopes(scopes.WithTenantSchema("tenant1")).Find(&Book{})
//	// SELECT * FROM tenant1.books;
//
// Example with Tabler interface:
//
//	type Book struct { ... }
//
//	func (u *Book) TableName() string { return "books" } // implements Tabler interface, no need to set TableName manually
//
//	db.Scopes(scopes.WithTenantSchema("tenant2")).Find(&Book{})
//	// SELECT * FROM tenant2.books;
//
// Example with model set manually:
//
//	type Book struct { ... }
//
//	db.Model(&Book{}).Scopes(scopes.WithTenantSchema("tenant3")).Find(&Book{}) // model is set manually.
//	// SELECT * FROM tenant3.books;
func WithTenantSchema(tenant string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var (
			tableName string
		)
		switch {
		case db.Statement.Table != "":
			// if the table name is set manually, use it
			tableName = db.Statement.Table
		case db.Statement.Model != nil:
			// if the table name is not set manually, try to get it from the model
			tableName = tableNameFromReflectValue(db.Statement.Model)
		case db.Statement.Dest != nil:
			// if the table name is not set manually, try to get it from the model
			tableName = tableNameFromReflectValue(db.Statement.Dest)
		}

		if tableName != "" {
			return db.Table(tenant + "." + tableName)
		}
		// otherwise, return an error
		_ = db.AddError(gorm.ErrModelValueRequired)
		return db
	}
}

// getTableName returns the table name from the model.
func getTableName(val interface{}) (string, bool) {
	if s, ok := val.(schema.Tabler); ok {
		return s.TableName(), true
	}
	return "", false
}

var (
	tableNameCache = make(map[string]string)
	cacheMutex     = &sync.RWMutex{}
)

func tableNameFromReflectValue(val interface{}) string {
	// Check if val implements the gorm.Tabler interface
	if tabler, ok := val.(schema.Tabler); ok {
		return tabler.TableName()
	}

	// Otherwise, use reflection
	value := reflect.Indirect(reflect.ValueOf(val))

	// Generate a key for the cache
	key := value.Type().PkgPath() + "." + value.Type().Name()

	// Check if the table name is in the cache
	cacheMutex.RLock()
	tableName, ok := tableNameCache[key]
	cacheMutex.RUnlock()

	if ok {
		// If the table name is in the cache, return it
		return tableName
	}

	// Otherwise, calculate the table name
	var name string
	switch value.Kind() { //nolint:exhaustive // this function is only concerned with structs and slices, no need for default case.
	case reflect.Struct:
		newElem := reflect.New(value.Type()).Interface()
		if n, ok := getTableName(newElem); ok {
			name = n
		}
	case reflect.Slice:
		elemType := value.Type().Elem()
		newElem := reflect.New(elemType).Interface()
		if n, ok := getTableName(newElem); ok {
			name = n
		}
	}

	if name != "" {
		// Store the table name in the cache
		cacheMutex.Lock()
		tableNameCache[key] = name
		cacheMutex.Unlock()
	}

	return name
}
