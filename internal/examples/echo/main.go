package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	multitenancy "github.com/bartventer/gorm-multitenancy/v2"
	"github.com/bartventer/gorm-multitenancy/v2/drivers/postgres"
	echomw "github.com/bartventer/gorm-multitenancy/v2/middleware/echo"
	"github.com/bartventer/gorm-multitenancy/v2/scopes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

const (
	// TableNameTenant is the table name for the tenant model
	TableNameTenant = "public.tenants"
	// TableNameBook is the table name for the book model
	TableNameBook = "books"
)

// TableName overrides the table name used by Tenant to `tenants`
func (u *Tenant) TableName() string { return TableNameTenant }

// Tenant is the tenant model
type Tenant struct {
	gorm.Model
	postgres.TenantModel
}

// Book is the book model
type Book struct {
	ID           uint   `gorm:"primarykey" json:"id"`
	Name         string `json:"name"`
	TenantSchema string `gorm:"column:tenant_schema"`
	Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
}

var _ multitenancy.TenantTabler = (*Book)(nil)

// TableName overrides the table name used by Book to `books`
func (u *Book) TableName() string { return TableNameBook }

// IsTenantTable returns true
func (u *Book) IsTenantTable() bool { return true }

type (
	// CreateTenantBody is the request body for creating a tenant
	CreateTenantBody struct {
		DomainURL string `json:"domainUrl"`
	}

	// UpdateBookBody is the request body for updating a book
	UpdateBookBody struct {
		Name string `json:"name"`
	}

	// BookResponse is the response body for a book
	BookResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	// TenantResponse is the response body for a tenant
	TenantResponse struct {
		ID        uint   `json:"id"`
		DomainURL string `json:"domainUrl"`
	}
)

// create database connection, models, and tables
func init() {
	var err error

	db, err = gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic(err)
	}

	// register models
	if err := postgres.RegisterModels(db, &Tenant{}, &Book{}); err != nil {
		panic(err)
	}

	// create models

	// create public schema
	if err := postgres.MigratePublicSchema(db); err != nil {
		panic(err)
	}

	// create tenants
	tenants := []Tenant{
		{
			TenantModel: postgres.TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "tenant1",
			},
		},
		{
			TenantModel: postgres.TenantModel{
				DomainURL:  "tenant2.example.com",
				SchemaName: "tenant2",
			},
		},
	}
	for _, tenant := range tenants {
		if err := db.Where("domain_url = ?", tenant.DomainURL).FirstOrCreate(&tenant).Error; err != nil {
			panic(err)
		}
	}

	// create schemas for tenants, and migrate "private" tables
	for _, tenant := range tenants {
		postgres.CreateSchemaForTenant(db, tenant.SchemaName)
	}

	// Create schema specific data
	books := []Book{{Name: "Book 1"}, {Name: "Book 2"}}
	db.Transaction(func(tx *gorm.DB) error {
		for _, tenant := range tenants {
			// set search path to tenant
			tx.Exec(fmt.Sprintf("SET search_path TO %s", tenant.SchemaName))
			for _, book := range books {
				book.Tenant = tenant
				book.Name = fmt.Sprintf("%s - %s", tenant.SchemaName, book.Name)
				if err := tx.Where("name = ?", book.Name).FirstOrCreate(&book).Error; err != nil {
					return err
				}
			}
			// Reset search path
			tx.Exec(fmt.Sprintf("SET search_path TO %s", "public"))
		}
		return nil
	})
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// create tenant middleware
	mw := echomw.WithTenant(echomw.WithTenantConfig{
		DB: db,
		Skipper: func(r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, "/tenants") // skip tenant routes
		},
		TenantGetters: echomw.DefaultTenantGetters,
	})
	e.Use(mw)

	// routes
	e.POST("/tenants", createTenantHandler)
	e.GET("/tenants/:id", getTenantHandler)
	e.DELETE("/tenants/:id", deleteTenantHandler)

	e.GET("/books", getBooksHandler)
	e.POST("/books", createBookHandler)
	e.DELETE("/books/:id", deleteBookHandler)
	e.PUT("/books/:id", updateBookHandler)

	// start echo server
	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}

func createTenantHandler(c echo.Context) error {
	var body CreateTenantBody
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	tenant := &Tenant{
		TenantModel: postgres.TenantModel{
			DomainURL:  body.DomainURL,
			SchemaName: strings.Split(body.DomainURL, ".")[0],
		},
	}

	// create tenant
	if err := db.Create(tenant).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// create schema for tenant, and migrate "private" tables
	if err := postgres.CreateSchemaForTenant(db, tenant.SchemaName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := &TenantResponse{
		ID:        tenant.ID,
		DomainURL: tenant.DomainURL,
	}
	return c.JSON(http.StatusCreated, res)
}

func getTenantHandler(c echo.Context) error {
	id := c.Param("id")
	// get tenant
	tenant := &TenantResponse{}
	if err := db.Table(TableNameTenant).First(tenant, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, tenant)
}

func deleteTenantHandler(c echo.Context) error {
	id := c.Param("id")
	// get tenant
	tenant := &Tenant{}
	if err := db.First(tenant, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	// delete schema for tenant
	if err := postgres.DropSchemaForTenant(db, tenant.SchemaName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// delete tenant
	if err := db.Delete(&Tenant{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func getBooksHandler(c echo.Context) error {
	tenant, err := echomw.TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	var books []BookResponse
	if err := db.Table(TableNameBook).Scopes(scopes.WithTenantSchema(tenant)).Find(&books).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, books)
}

func createBookHandler(c echo.Context) error {
	tenant, err := echomw.TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	var book Book
	if err := c.Bind(&book); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	book.TenantSchema = tenant
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).Create(&book).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := &BookResponse{
		ID:   book.ID,
		Name: book.Name,
	}
	return c.JSON(http.StatusCreated, res)
}

func deleteBookHandler(c echo.Context) error {
	tenant, err := echomw.TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	id := c.Param("id")
	// get book
	var book Book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).First(&book, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	// delete book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).Delete(&Book{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func updateBookHandler(c echo.Context) error {
	tenant, err := echomw.TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	id := c.Param("id")
	var body UpdateBookBody
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// get book
	book := &Book{}
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).First(book, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	// update book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).Model(book).Where("id = ?", id).Updates(body).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
