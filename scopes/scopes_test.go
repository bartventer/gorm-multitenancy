package scopes

import (
	"regexp"
	"strings"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v6/internal/testutil"
	"gorm.io/gorm"
)

var DB = testutil.NewTestDB()

type Book struct {
	ID    uint
	Title string
}

func (Book) TableName() string {
	return "books"
}

// assertEqualSQL for assert that the sql is equal, this method will ignore quote, and dialect specials.
func assertEqualSQL(t *testing.T, expected string, actually string) {
	t.Helper()

	// replace SQL quote, convert into postgresql like ""
	expected = replaceQuoteInSQL(expected)
	actually = replaceQuoteInSQL(actually)

	// ignore updated_at value, because it's generated in Gorm internal, can't to mock value on update.
	updatedAtRe := regexp.MustCompile(`(?i)"updated_at"=".+?"`)
	actually = updatedAtRe.ReplaceAllString(actually, `"updated_at"=?`)
	expected = updatedAtRe.ReplaceAllString(expected, `"updated_at"=?`)

	// ignore RETURNING "id" (only in PostgreSQL)
	returningRe := regexp.MustCompile(`(?i)RETURNING "id"`)
	actually = returningRe.ReplaceAllString(actually, ``)
	expected = returningRe.ReplaceAllString(expected, ``)

	actually = strings.TrimSpace(actually)
	expected = strings.TrimSpace(expected)

	if actually != expected {
		t.Fatalf("\nexpected: %s\nactually: %s", expected, actually)
	}
}

func replaceQuoteInSQL(sql string) string {
	// convert single quote into double quote
	sql = strings.ReplaceAll(sql, `'`, `"`)

	// convert dialect special quote into double quote
	switch DB.Dialector.Name() {
	case "postgres":
		sql = strings.ReplaceAll(sql, `"`, `"`)
	case "mysql", "sqlite":
		sql = strings.ReplaceAll(sql, "`", `"`)
	case "sqlserver":
		sql = strings.ReplaceAll(sql, `'`, `"`)
	}

	return sql
}

func TestWithTenantSchema(t *testing.T) {
	tests := []struct {
		name     string
		queryFn  func(tx *gorm.DB) *gorm.DB
		expected string
	}{
		{
			name: "valid: with table name set",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Table("books").Scopes(WithTenantSchema("tenant2")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant2"."books"`,
		},
		{
			name: "valid: with model set",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Model(&Book{}).Scopes(WithTenantSchema("tenant1")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "valid: with dest pointer to struct",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "invalid: dest not a pointer",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(Book{})
			},
			expected: ``,
		},
		{
			name: "valid: with dest pointer to array/slice",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(&[]Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "invalid: Tabler interface not implemented",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant3")).Find(&struct{}{})
			},
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqualSQL(t, tt.expected, DB.ToSQL(tt.queryFn))
		})
	}
}
