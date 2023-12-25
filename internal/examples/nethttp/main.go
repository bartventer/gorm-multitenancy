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
	// TableNameTenant is the table name for the tenant model
	TableNameTenant = "tenants"
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
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `json:"name"`
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
