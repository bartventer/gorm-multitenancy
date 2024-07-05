package drivertest

import (
	"testing"
	"time"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"gorm.io/gorm"
)

// Valid models for testing.
type (
	// the main shared model. user will be our tenant.
	userShared struct {
		// check constraint added to test this error is fixed on mysql:
		//  Error 3822 (HY000): Duplicate check constraint name '...'
		ID        string `gorm:"primaryKey;size:63;check:LENGTH(id) >= 3;"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	// tenant-specific model, with:
	//   - BelongsTo association to shared model (e.g. author belongs to user)
	//   - HasMany association to tenant-specific model (e.g. author has many books)
	//   - ManyToMany association to tenant-specific model (e.g. author has many languages/vice-versa)
	authorPrivate struct {
		gorm.Model
		User   userShared `gorm:"foreignKey:UserID"`
		UserID string
		Books  []*bookPrivate `gorm:"foreignKey:AuthorID"`
	}

	// the shared model for the many-to-many association.
	languagePrivate struct {
		gorm.Model
		Name  string
		Books []*bookPrivate `gorm:"many2many:book_languages;"`
	}

	// tenant-specific model, with:
	//   - BelongsTo association to tenant-specific model (e.g. book belongs to author)
	//   - ManyToMany association to shared model (e.g. book has many languages)
	bookPrivate struct {
		gorm.Model
		Title     string
		AuthorID  uint               `gorm:"foreignKey:ID"`
		Languages []*languagePrivate `gorm:"many2many:book_languages;"`
	}
)

var _ driver.TenantTabler = new(userShared)
var _ driver.TenantTabler = new(authorPrivate)
var _ driver.TenantTabler = new(languagePrivate)
var _ driver.TenantTabler = new(bookPrivate)

func (userShared) TableName() string   { return "public.users" }
func (userShared) IsSharedModel() bool { return true }

func (authorPrivate) TableName() string   { return "authors" }
func (authorPrivate) IsSharedModel() bool { return false }

func (languagePrivate) TableName() string   { return "languages" }
func (languagePrivate) IsSharedModel() bool { return false }

func (bookPrivate) TableName() string   { return "books" }
func (bookPrivate) IsSharedModel() bool { return false }

// makeAllModels returns all valid models for testing.
func makeAllModels[TB testing.TB](t TB) []driver.TenantTabler {
	t.Helper()
	return []driver.TenantTabler{
		// Shared
		&userShared{},
		// Private
		&authorPrivate{},
		&bookPrivate{},
		&languagePrivate{},
	}
}

// makeSharedModels returns all valid shared models for testing.
func makeSharedModels[TB testing.TB](t TB) []driver.TenantTabler {
	t.Helper()
	all := makeAllModels(t)
	out := make([]driver.TenantTabler, 0, len(all))
	for _, m := range all {
		if m.IsSharedModel() {
			out = append(out, m)
		}
	}
	return out
}

// makePrivateModels returns all valid tenant-specific models for testing.
func makePrivateModels[TB testing.TB](t TB) []driver.TenantTabler {
	t.Helper()
	all := makeAllModels(t)
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
	userSharedInvalid  struct{} // invalid shared model.
	bookPrivateInvalid struct{} // invalid tenant-specific model.
)

// Invalid models for testing.
var _ driver.TenantTabler = new(userSharedInvalid)
var _ driver.TenantTabler = new(bookPrivateInvalid)

func (userSharedInvalid) TableName() string   { return "users" } // missing .public prefix
func (userSharedInvalid) IsSharedModel() bool { return true }

func (bookPrivateInvalid) TableName() string   { return "public.books" } // contains .public prefix
func (bookPrivateInvalid) IsSharedModel() bool { return false }

type mockTenantModel struct {
	gorm.Model
	multitenancy.TenantModel
}

var _ driver.TenantTabler = new(mockTenantModel)

func (mockTenantModel) TableName() string   { return "public.tenants" }
func (mockTenantModel) IsSharedModel() bool { return true }
