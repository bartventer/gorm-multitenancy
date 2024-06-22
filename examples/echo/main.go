package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	echomw "github.com/bartventer/gorm-multitenancy/middleware/echo/v7"
	"github.com/bartventer/gorm-multitenancy/postgres/v7"
	"github.com/bartventer/gorm-multitenancy/postgres/v7/scopes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	db *gorm.DB
)

const (
	// TableNameTenant is the table name for the tenant model.
	TableNameTenant = "public.tenants"
	// TableNameBook is the table name for the book model.
	TableNameBook = "books"
)

// TableName overrides the table name used by Tenant to `tenants`.
func (Tenant) TableName() string { return TableNameTenant }

// Tenant is the tenant model.
type Tenant struct {
	gorm.Model
	postgres.TenantModel
}

// Book is the book model.
type Book struct {
	ID           uint   `json:"id"   gorm:"primarykey"`
	Name         string `json:"name" gorm:"column:name;size:255;not null;default:NULL"`
	TenantSchema string `            gorm:"column:tenant_schema"`
	Tenant       Tenant `            gorm:"foreignKey:TenantSchema;references:SchemaName"`
}

var _ postgres.TenantTabler = (*Book)(nil)

// TableName overrides the table name used by Book to `books`.
func (Book) TableName() string { return TableNameBook }

// IsTenantTable returns true.
func (Book) IsTenantTable() bool { return true }

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

// create database connection, models, and tables.
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
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "schema_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"domain_url"}),
		}).Model(&tenant).Where("schema_name = ?", tenant.SchemaName).FirstOrCreate(&tenant).Error; err != nil {
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

// TenantFromContext returns the tenant from the context.
func TenantFromContext(c echo.Context) (string, error) {
	tenant, ok := c.Get(echomw.TenantKey.String()).(string)
	if !ok {
		return "", fmt.Errorf("no tenant in context")
	}
	return tenant, nil
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// create tenant middleware
	mw := echomw.WithTenant(echomw.WithTenantConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/tenants") // skip tenant routes
		},
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
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, tx.Error.Error())
	}
	// always rollback the transaction
	defer tx.Rollback()
	// create tenant
	if err := tx.Create(tenant).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// create schema for tenant, and migrate "private" tables
	if err := postgres.CreateSchemaForTenant(tx, tenant.SchemaName); err != nil {
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
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, tx.Error.Error())
	}
	// always rollback the transaction
	defer tx.Rollback()
	// delete schema for tenant
	if err := postgres.DropSchemaForTenant(tx, tenant.SchemaName); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// delete tenant
	if err := tx.Delete(&Tenant{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func getBooksHandler(c echo.Context) error {
	tenant, err := TenantFromContext(c)
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
	tenant, err := TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	var book Book
	if err := c.Bind(&book); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	book.TenantSchema = tenant
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, tx.Error.Error())
	}
	// always rollback the transaction
	defer tx.Rollback()
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).Create(&book).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := &BookResponse{
		ID:   book.ID,
		Name: book.Name,
	}
	return c.JSON(http.StatusCreated, res)
}

func deleteBookHandler(c echo.Context) error {
	tenant, err := TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	id := c.Param("id")
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, tx.Error.Error())
	}
	// always rollback the transaction
	defer tx.Rollback()
	// get book
	var book Book
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).First(&book, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	// delete book
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).Delete(&Book{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func updateBookHandler(c echo.Context) error {
	tenant, err := TenantFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	id := c.Param("id")
	var body UpdateBookBody
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if body.Name == "" { // this is ugly but just for the sake of the example
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, tx.Error.Error())
	}
	// always rollback the transaction
	defer tx.Rollback()
	// get book
	book := &Book{}
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).First(book, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	// update book
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).Model(book).Updates(Book{
		Name: body.Name,
	}).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
