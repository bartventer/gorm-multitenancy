package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
	"github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/scopes"
	nethttpmw "github.com/bartventer/gorm-multitenancy/middleware/nethttp/v7"
	"github.com/urfave/negroni"
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

// TenantFromContext returns the tenant from the given HTTP request's context.
func TenantFromContext(ctx context.Context) (string, error) {
	tenant, ok := ctx.Value(nethttpmw.TenantKey).(string)
	if !ok {
		return "", fmt.Errorf("failed to get tenant from context")
	}
	return tenant, nil
}

func main() {
	mux := http.NewServeMux()

	// routes
	mux.HandleFunc("POST /tenants", createTenantHandler)
	mux.HandleFunc("GET /tenants/{id}", getTenantHandler)
	mux.HandleFunc("DELETE /tenants/{id}", deleteTenantHandler)
	mux.HandleFunc("POST /books", createBookHandler)
	mux.HandleFunc("GET /books", getBooksHandler)
	mux.HandleFunc("DELETE /books/{id}", deleteBookHandler)
	mux.HandleFunc("PUT /books/{id}", updateBookHandler)

	// Global middleware
	tenantMux := nethttpmw.WithTenant(nethttpmw.WithTenantConfig{
		Skipper: func(r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, "/tenants")
		},
	})(mux)
	n := negroni.Classic()
	n.UseHandler(tenantMux)

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      n,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
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

	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	// create tenant
	if err := tx.Create(tenant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// create schema for tenant, and migrate "private" tables
	if err := postgres.CreateSchemaForTenant(tx, tenant.SchemaName); err != nil {
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
	id := r.PathValue("id")
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
	id := r.PathValue("id")
	// get tenant
	tenant := &Tenant{}
	if err := db.First(tenant, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	// always rollback the transaction
	defer tx.Rollback()
	// delete schema for tenant
	if err := postgres.DropSchemaForTenant(tx, tenant.SchemaName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// delete tenant
	if err := tx.Delete(&Tenant{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getBooksHandler(w http.ResponseWriter, r *http.Request) {
	tenant, err := TenantFromContext(r.Context())
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
	tenant, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	book.TenantSchema = tenant
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	// always rollback the transaction
	defer tx.Rollback()
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).Create(&book).Error; err != nil {
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
	tenant, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	// always rollback the transaction
	defer tx.Rollback()
	// get book
	var book Book
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).First(&book, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// delete book
	if err := tx.Scopes(scopes.WithTenantSchema(tenant)).Delete(&Book{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	tenant, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	var body UpdateBookBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Name == "" { // this is ugly but just for the sake of the example
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	// start transaction
	tx := db.Begin()
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	// always rollback the transaction
	defer tx.Rollback()
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
