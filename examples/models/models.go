package models

import (
	multitenancy "github.com/bartventer/gorm-multitenancy/v7"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"gorm.io/gorm"
)

const (
	TableNameTenant = "public.tenants" // TableNameTenant is the table name for the tenant model.
	TableNameBook   = "books"          // TableNameBook is the table name for the book model.
)

type (
	// Tenant is the tenant model.
	Tenant struct {
		gorm.Model
		multitenancy.TenantModel
	}

	// Book is the book model.
	Book struct {
		gorm.Model
		Name         string `gorm:"column:name;size:255;not null;"`
		TenantSchema string `gorm:"column:tenant_schema"`
		Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
	}
)

var _ driver.TenantTabler = new(Tenant)
var _ driver.TenantTabler = new(Book)

func (Tenant) TableName() string   { return TableNameTenant }
func (Tenant) IsSharedModel() bool { return true }

func (Book) TableName() string   { return TableNameBook }
func (Book) IsSharedModel() bool { return false }

type (
	// CreateTenantBody is the request body for creating a tenant.
	CreateTenantBody struct {
		DomainURL string `json:"domainUrl"`
	}

	// UpdateBookBody is the request body for updating a book.
	UpdateBookBody struct {
		Name string `json:"name"`
	}

	// BookResponse is the response body for a book.
	BookResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	// TenantResponse is the response body for a tenant.
	TenantResponse struct {
		ID        uint   `json:"id"`
		DomainURL string `json:"domainUrl"`
	}
)
