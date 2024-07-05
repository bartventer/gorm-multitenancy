package scopes

import (
	"regexp"
	"strings"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
)

type Book struct {
	ID    uint
	Title string
}

var _ driver.TenantTabler = new(Book)

func (Book) TableName() string   { return "books" }
func (Book) IsSharedModel() bool { return true }

// assertEqualSQL for assert that the sql is equal, this method will ignore quote, and dialect specials.
func assertEqualSQL(t *testing.T, db *gorm.DB, expected string, actually string) {
	t.Helper()

	// replace SQL quote, convert into postgresql like ""
	expected = replaceQuoteInSQL(t, db, expected)
	actually = replaceQuoteInSQL(t, db, actually)

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

func replaceQuoteInSQL(t *testing.T, db *gorm.DB, sql string) string {
	t.Helper()
	// convert single quote into double quote
	sql = strings.ReplaceAll(sql, `'`, `"`)

	// convert dialect special quote into double quote
	switch db.Name() {
	case "postgres":
		sql = strings.ReplaceAll(sql, `"`, `"`)
	case "mysql", "sqlite":
		sql = strings.ReplaceAll(sql, "`", `"`)
	case "sqlserver":
		sql = strings.ReplaceAll(sql, `'`, `"`)
	case "dummy": //See dummy_dialecter.go
		sql = strings.ReplaceAll(sql, "`", `"`)
	}

	return sql
}

func TestWithTenantSchema(t *testing.T) {
	db, err := gorm.Open(tests.DummyDialector{
		TranslatedErr: nil,
	})
	if err != nil {
		t.Fatalf("failed to open database connection: %v", err)
	}
	tests := []struct {
		name     string
		queryFn  func(tx *gorm.DB) *gorm.DB
		expected string
	}{
		{
			name: "01 - With table name set",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Table("books").Scopes(WithTenantSchema("tenant2")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant2"."books"`,
		},
		{
			name: "02 - With model set",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Model(Book{}).Scopes(WithTenantSchema("tenant1")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "03 - With model (pointer) set",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Model(&Book{}).Scopes(WithTenantSchema("tenant1")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "04 - With dest pointer to struct",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "05 - With dest struct",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "06 - With dest pointer to array/slice",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(&[]Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "07 - With dest pointer to array/slice non empty",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find(&[]Book{
					{ID: 1, Title: "Book 1"},
				})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "08 - With dest array/slice",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find([]Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "09 - With dest array/slice of pointer to structs",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant1")).Find([]*Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "10 - Invalid: Tabler interface not implemented",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant3")).Find(&struct{}{})
			},
			expected: ``,
		},
		{
			name: "11 - Invalid: Tabler interface not implemented (slice/array)",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Scopes(WithTenantSchema("tenant3")).Find(&[]struct{}{})
			},
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqualSQL(t, db, tt.expected, db.ToSQL(tt.queryFn))
		})
	}
}

func Test_tableNameFromInterface(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "invalid: nil",
			args: args{i: nil},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tableNameFromInterface(tt.args.i); got != tt.want {
				t.Errorf("tableNameFromInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}
