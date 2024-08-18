// Package testmodels provides valid and invalid models for testing.
package testmodels

import (
	"testing"
	"time"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"gorm.io/gorm"
)

// Valid models for testing.
type (
	Tenant struct {
		ID        string `gorm:"primaryKey;size:63;check:LENGTH(id) >= 3;"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	Author struct {
		gorm.Model
		Tenant   Tenant `gorm:"foreignKey:TenantID"`
		TenantID string
		Books    []*Book `gorm:"foreignKey:AuthorID"`
	}

	Language struct {
		gorm.Model
		Name  string
		Books []*Book `gorm:"many2many:book_languages;"`
	}

	Book struct {
		gorm.Model
		Title     string
		AuthorID  uint        `gorm:"index"`
		Languages []*Language `gorm:"many2many:book_languages;"`
	}
)

var _ driver.TenantTabler = new(Tenant)
var _ driver.TenantTabler = new(Author)
var _ driver.TenantTabler = new(Language)
var _ driver.TenantTabler = new(Book)

func (Tenant) TableName() string   { return "public.tenants" }
func (Tenant) IsSharedModel() bool { return true }

func (Author) TableName() string   { return "authors" }
func (Author) IsSharedModel() bool { return false }

func (Language) TableName() string   { return "languages" }
func (Language) IsSharedModel() bool { return false }

func (Book) TableName() string   { return "books" }
func (Book) IsSharedModel() bool { return false }

// MakeAllModels returns all valid models for testing.
func MakeAllModels[TB testing.TB](t TB) []driver.TenantTabler {
	t.Helper()
	return []driver.TenantTabler{
		// Shared
		&Tenant{},
		// Private
		&Author{},
		&Book{},
		&Language{},
	}
}

// MakeSharedModels returns all valid shared models for testing.
func MakeSharedModels[TB testing.TB](t TB) []driver.TenantTabler {
	t.Helper()
	all := MakeAllModels(t)
	out := make([]driver.TenantTabler, 0, len(all))
	for _, m := range all {
		if m.IsSharedModel() {
			out = append(out, m)
		}
	}
	return out
}

// MakePrivateModels returns all valid tenant-specific models for testing.
func MakePrivateModels[TB testing.TB](t TB) []driver.TenantTabler {
	t.Helper()
	all := MakeAllModels(t)
	out := make([]driver.TenantTabler, 0, len(all))
	for _, m := range all {
		if !m.IsSharedModel() {
			out = append(out, m)
		}
	}
	return out
}

// Invalid models for testing.
type (
	TenantInvalid struct{} // invalid shared model.
	BookInvalid   struct{} // invalid tenant-specific model.
)

// Invalid models for testing.
var _ driver.TenantTabler = new(TenantInvalid)
var _ driver.TenantTabler = new(BookInvalid)

func (TenantInvalid) TableName() string   { return "tenants" } // missing .public prefix
func (TenantInvalid) IsSharedModel() bool { return true }

func (BookInvalid) TableName() string   { return "public.books" } // contains .public prefix
func (BookInvalid) IsSharedModel() bool { return false }

// FakeTenant embeds [gorm.Model] and [multitenancy.TenantModel].
type FakeTenant struct {
	gorm.Model
	multitenancy.TenantModel
}

var _ driver.TenantTabler = new(FakeTenant)

func (FakeTenant) TableName() string   { return "public.tenants" }
func (FakeTenant) IsSharedModel() bool { return true }
