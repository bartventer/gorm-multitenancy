/*
Package scopes provides a set of predefined GORM scopes for managing multi-tenant applications using the gorm-multitenancy library.
*/
package scopes

import (
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
			tn string
		)
		switch {
		case db.Statement.Table != "":
			tn = db.Statement.Table
		case db.Statement.Model != nil:
			tn = tableNameFromReflectValue(db.Statement.Model)
		case db.Statement.Dest != nil:
			tn = tableNameFromReflectValue(db.Statement.Dest)
		}

		if tn != "" {
			return db.Table(tenant + "." + tn)
		}
		// otherwise, return an error
		_ = db.AddError(gorm.ErrModelValueRequired)
		return db
	}
}

// tableNameFromReflectValue returns the table name from the model.
func tableNameFromReflectValue(val interface{}) string {
	if s, ok := val.(schema.Tabler); ok {
		return s.TableName()
	}
	return ""
}
