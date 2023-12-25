# gorm-multitenancy

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gorm-multitenancy)](https://goreportcard.com/report/github.com/bartventer/gorm-multitenancy)
[![Coverage Status](https://coveralls.io/repos/github/bartventer/gorm-multitenancy/badge.svg?branch=master)](https://coveralls.io/github/bartventer/gorm-multitenancy?branch=master)
[![Build](https://github.com/bartventer/gorm-multitenancy/actions/workflows/go.yml/badge.svg)](https://github.com/bartventer/gorm-multitenancy/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/bartventer/gorm-multitenancy.svg)](LICENSE)

There are three common approaches to multitenancy in a database:
- Shared database, shared schema
- Shared database, separate schemas
- Separate databases

This package implements the shared database, separate schemas approach. It uses the [gorm](https://gorm.io/) ORM to manage the database and provides custom drivers to support multitenancy. It also provides HTTP middleware to retrieve the tenant from the request and set the tenant in context.

## Database compatibility
Current supported databases are listed below. Pull requests for other drivers are welcome.
- [PostgreSQL](https://www.postgresql.org/)

## Router compatibility
Current supported routers are listed below. Pull requests for other routers are welcome.
- [echo](https://echo.labstack.com/docs)
- [net/http](https://golang.org/pkg/net/http/)

## Installation

```bash
go get -u github.com/bartventer/gorm-multitenancy
```

## Examples

### [PostgreSQL](https://www.postgresql.org/) driver and [echo](https://echo.labstack.com/docs) middleware

```go
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	multitenancy "github.com/bartventer/gorm-multitenancy"
	"github.com/bartventer/gorm-multitenancy/drivers/postgres"
	echomw "github.com/bartventer/gorm-multitenancy/middleware/echo"
	"github.com/bartventer/gorm-multitenancy/scopes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

const (
	TableNameTenant = "tenants"
	TableNameBook   = "books"
)

func (u *Tenant) TableName() string { return TableNameTenant }

type Tenant struct {
	gorm.Model
	postgres.TenantModel
}
type Book struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `json:"name"`
}

var _ multitenancy.TenantTabler = (*Book)(nil)

func (u *Book) TableName() string { return TableNameBook }

func (u *Book) IsTenantTable() bool { return true }

type (
	CreateTenantBody struct {
		DomainURL string `json:"domainUrl"`
	}

	UpdateBookBody struct {
		Name string `json:"name"`
	}

	BookResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

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
	postgres.RegisterModels(db, &Tenant{}, &Book{})

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

	// Create data for tenant1 (private schema)
	books := []Book{{Name: "Book 1"}, {Name: "Book 2"}}
	db.Transaction(func(tx *gorm.DB) error {
		// set search path to tenant
		tx.Exec(fmt.Sprintf("SET search_path TO %s", tenants[0].SchemaName))
		for _, book := range books {
			if err := tx.Where("name = ?", book.Name).FirstOrCreate(&book).Error; err != nil {
				return err
			}
		}
		// Reset search path
		tx.Exec(fmt.Sprintf("SET search_path TO %s", "public"))
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
	id, _ := strconv.Atoi(c.Param("id"))
	// get tenant
	tenant := &TenantResponse{}
	if err := db.Table(TableNameTenant).First(tenant, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, tenant)
}

func deleteTenantHandler(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
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
	id, _ := strconv.Atoi(c.Param("id"))
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

```

### [PostgreSQL](https://www.postgresql.org/) driver and [net/http](https://golang.org/pkg/net/http/) middleware

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	multitenancy "github.com/bartventer/gorm-multitenancy"
	"github.com/bartventer/gorm-multitenancy/drivers/postgres"
	nethttpmw "github.com/bartventer/gorm-multitenancy/middleware/nethttp"
	"github.com/bartventer/gorm-multitenancy/scopes"
	"github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

const (
	TableNameTenant = "tenants"
	TableNameBook   = "books"
)

func (u *Tenant) TableName() string { return TableNameTenant }

type Tenant struct {
	gorm.Model
	postgres.TenantModel
}
type Book struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `json:"name"`
}

var _ multitenancy.TenantTabler = (*Book)(nil)

func (u *Book) TableName() string { return TableNameBook }

func (u *Book) IsTenantTable() bool { return true }

type (
	CreateTenantBody struct {
		DomainURL string `json:"domainUrl"`
	}

	UpdateBookBody struct {
		Name string `json:"name"`
	}

	BookResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

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
	postgres.RegisterModels(db, &Tenant{}, &Book{})

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

	// Create data for tenant1 (private schema)
	books := []Book{{Name: "Book 1"}, {Name: "Book 2"}}
	db.Transaction(func(tx *gorm.DB) error {
		// set search path to tenant
		tx.Exec(fmt.Sprintf("SET search_path TO %s", tenants[0].SchemaName))
		for _, book := range books {
			if err := tx.Where("name = ?", book.Name).FirstOrCreate(&book).Error; err != nil {
				return err
			}
		}
		// Reset search path
		tx.Exec(fmt.Sprintf("SET search_path TO %s", "public"))
		return nil
	})
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// create tenant middleware
	mw := nethttpmw.WithTenant(nethttpmw.WithTenantConfig{
		DB: db,
		Skipper: func(r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, "/tenants") // skip tenant routes
		},
		TenantGetters: nethttpmw.DefaultTenantGetters,
	})

	r.Use(mw)

	// routes
	r.Post("/tenants", createTenantHandler)
	r.Get("/tenants/{id}", getTenantHandler)
	r.Delete("/tenants/{id}", deleteTenantHandler)

	r.Post("/books", createBookHandler)
	r.Get("/books", getBooksHandler)
	r.Delete("/books/{id}", deleteBookHandler)
	r.Put("/books/{id}", updateBookHandler)

	// start chi server
	http.ListenAndServe(":8080", r)
}

func createTenantHandler(w http.ResponseWriter, r *http.Request) {
	var body CreateTenantBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tenant := &Tenant{
		TenantModel: postgres.TenantModel{
			DomainURL:  body.DomainURL,
			SchemaName: strings.Split(body.DomainURL, ".")[0],
		},
	}

	// create tenant
	if err := db.Create(tenant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// create schema for tenant, and migrate "private" tables
	if err := postgres.CreateSchemaForTenant(db, tenant.SchemaName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	res := &TenantResponse{
		ID:        tenant.ID,
		DomainURL: tenant.DomainURL,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	tenant := &TenantResponse{}
	if err := db.Table(TableNameTenant).First(tenant, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(tenant); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func deleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	// get tenant
	tenant := &Tenant{}
	if err := db.First(tenant, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// delete schema for tenant
	if err := postgres.DropSchemaForTenant(db, tenant.SchemaName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// delete tenant
	if err := db.Delete(&Tenant{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getBooksHandler(w http.ResponseWriter, r *http.Request) {
	tenant, err := nethttpmw.TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var books []BookResponse
	if err := db.Table(TableNameBook).Scopes(scopes.WithTenantSchema(tenant)).Find(&books).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createBookHandler(w http.ResponseWriter, r *http.Request) {
	tenant, err := nethttpmw.TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).Create(&book).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := &BookResponse{
		ID:   book.ID,
		Name: book.Name,
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	tenant, err := nethttpmw.TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	// get book
	var book Book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).First(&book, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// delete book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).Delete(&Book{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	tenant, err := nethttpmw.TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var body UpdateBookBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// get book
	var book Book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).First(&book, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// update book
	if err := db.Scopes(scopes.WithTenantSchema(tenant)).Model(&book).Updates(body).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

```

### Example usage

#### Create tenant
 - Parse the request body into a CreateTenantBody struct
 - Create the tenant in the database (public schema)
 - Create the schema for the tenant
 - Return the HTTP status code 201 and the tenant in the response body

##### Request
```bash
curl -X POST \
  http://example.com:8080/tenants \
  -H 'Content-Type: application/json' \
  -d '{
  "domainUrl": "tenant3.example.com"
}'
```

##### Response
```json
{
  "id": 3,
  "domainUrl": "tenant3.example.com"
}
```

#### Get tenant
 - Get the tenant from the database
 - Return the HTTP status code 200 and the tenant in the response body

##### Request
```bash
curl -X GET \
  http://example.com:8080/tenants/3
```

##### Response
```json
{
  "id": 3,
  "domainUrl": "tenant3.example.com"
}
```

#### Delete tenant
 - Get the tenant from the database
 - Delete the schema for the tenant
 - Delete the tenant from the database
 - Return the HTTP status code 204

##### Request
```bash
curl -X DELETE \
  http://example.com:8080/tenants/3
```

##### Response
```json
```

#### Get books
 - Get the tenant from the request host or header
 - Get all books for the tenant
 - Return the HTTP status code 200 and the books in the response body

##### Request
```bash
curl -X GET \
  http://example.com:8080/books \
  -H 'Host: tenant1.example.com'
```

##### Response
```json
[
  {
    "id": 1,
    "name": "Book 1"
  },
  {
    "id": 2,
    "name": "Book 2"
  }
]
```

#### Create book
 - Get the tenant from the request host or header
 - Parse the request body into a Book struct
 - Create the book for the tenant in the database
 - Return the HTTP status code 201 and the book in the response body

##### Request
```bash
curl -X POST \
  http://example.com:8080/books \
  -H 'Content-Type: application/json' \
  -H 'Host: tenant1.example.com' \
  -d '{
  "name": "Book 3"
}'
```

##### Response
```json
{
  "id": 3,
  "name": "Book 3"
}
```

#### Delete book
 - Get the tenant from the request host or header
 - Get the book from the database
 - Delete the book from the database
 - Return the HTTP status code 204

##### Request
```bash
curl -X DELETE \
  http://example.com:8080/books/3 \
  -H 'Host: tenant1.example.com'
```

##### Response
```json
```

#### Update book
 - Get the tenant from the request host or header
 - Get the book from the database
 - Parse the request body into a UpdateBookBody struct
 - Update the book in the database
 - Return the HTTP status code 200

##### Request
```bash
curl -X PUT \
  http://example.com:8080/books/2 \
  -H 'Content-Type: application/json' \
  -H 'Host: tenant1.example.com' \
  -d '{
  "name": "Book 2 updated"
}'
```

##### Response
```json
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.