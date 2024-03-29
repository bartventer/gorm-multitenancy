package scopes

import (
	"regexp"
	"strings"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v5/internal"
	"gorm.io/gorm"
)

var DB = internal.NewTestDB()

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
			name: "test-with-tenant-schema",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Model(&Book{}).Scopes(WithTenantSchema("tenant1")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant1"."books"`,
		},
		{
			name: "test-with-tenant-schema-and-table-set-manually",
			queryFn: func(tx *gorm.DB) *gorm.DB {
				return tx.Table("books").Scopes(WithTenantSchema("tenant2")).Find(&Book{})
			},
			expected: `SELECT * FROM "tenant2"."books"`,
		},
		{
			name: "invalid:table-name-not-set",
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

func Test_getTableName(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "Value with TableName method",
			args: args{
				val: &Book{},
			},
			want:  "books",
			want1: true,
		},
		{
			name: "Value without TableName method",
			args: args{
				val: &struct{}{},
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getTableName(tt.args.val)
			if got != tt.want {
				t.Errorf("getTableName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getTableName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_tableNameFromReflectValue(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Struct with TableName method",
			args: args{
				val: &Book{},
			},
			want: "books",
		},
		{
			name: "Slice of struct with TableName method",
			args: args{
				val: &[]Book{},
			},
			want: "books",
		},
		{
			name: "Struct without TableName method",
			args: args{
				val: &struct{}{},
			},
			want: "",
		},
		{
			name: "Slice of struct without TableName method",
			args: args{
				val: &[]struct{}{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tableNameFromReflectValue(tt.args.val); got != tt.want {
				t.Errorf("tableNameFromReflectValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
