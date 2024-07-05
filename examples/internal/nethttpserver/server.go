package nethttpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/models"
	nethttpmw "github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"

	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/scopes"
	"github.com/urfave/negroni"
)

// Start starts the HTTP server.
func Start(db *multitenancy.DB) {
	cr := &controller{db: db}
	cr.start()
}

type controller struct {
	db   *multitenancy.DB
	once sync.Once
}

func (cr *controller) start() {
	cr.once.Do(func() {
		mux := http.NewServeMux()

		// routes
		mux.HandleFunc("POST /tenants", cr.createTenantHandler)
		mux.HandleFunc("GET /tenants/{id}", cr.getTenantHandler)
		mux.HandleFunc("DELETE /tenants/{id}", cr.deleteTenantHandler)
		mux.HandleFunc("POST /books", cr.createBookHandler)
		mux.HandleFunc("GET /books", cr.getBooksHandler)
		mux.HandleFunc("DELETE /books/{id}", cr.deleteBookHandler)
		mux.HandleFunc("PUT /books/{id}", cr.updateBookHandler)

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
	})
}

// TenantFromContext returns the tenant from the given HTTP request's context.
func TenantFromContext(ctx context.Context) (string, error) {
	tenant, ok := ctx.Value(nethttpmw.TenantKey).(string)
	if !ok {
		return "", errors.New("failed to get tenant from context")
	}
	return tenant, nil
}

func (cr *controller) createTenantHandler(w http.ResponseWriter, r *http.Request) {
	var body models.CreateTenantBody
	var err error
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	subdomain, subdomainErr := nethttpmw.ExtractSubdomain(body.DomainURL)
	if subdomainErr != nil {
		http.Error(w, subdomainErr.Error(), http.StatusBadRequest)
		return
	}
	tenant := &models.Tenant{
		TenantModel: multitenancy.TenantModel{
			DomainURL:  body.DomainURL,
			SchemaName: subdomain,
		},
	}
	if err = cr.db.Create(tenant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = cr.db.MigrateTenantModels(context.Background(), tenant.SchemaName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	res := &models.TenantResponse{
		ID:        tenant.ID,
		DomainURL: tenant.DomainURL,
	}
	if err = json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cr *controller) getTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.PathValue("id")
	tenant := &models.TenantResponse{}
	var err error
	if err = cr.db.Table(models.TableNameTenant).First(tenant, tenantID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err = json.NewEncoder(w).Encode(tenant); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cr *controller) deleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := r.PathValue("id")
	tenant := &models.Tenant{}
	var err error
	if err = cr.db.First(tenant, tenantID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err = cr.db.OffboardTenant(context.Background(), tenant.SchemaName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = cr.db.Delete(&models.Tenant{}, tenantID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cr *controller) getBooksHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var books []models.BookResponse
	if err = cr.db.Table(models.TableNameBook).Scopes(scopes.WithTenantSchema(tenantID)).Find(&books).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cr *controller) createBookHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var book models.Book
	if err = json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	book.TenantSchema = tenantID

	reset, tenantErr := cr.db.UseTenant(context.Background(), tenantID)
	if tenantErr != nil {
		http.Error(w, tenantErr.Error(), http.StatusInternalServerError)
		return
	}
	defer reset()
	if err = cr.db.Create(&book).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := &models.BookResponse{
		ID:   book.ID,
		Name: book.Name,
	}
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cr *controller) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bookID := r.PathValue("id")

	var book models.Book
	if err = cr.db.Scopes(scopes.WithTenantSchema(tenantID)).First(&book, bookID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err = cr.db.Scopes(scopes.WithTenantSchema(tenantID)).Delete(&models.Book{}, bookID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cr *controller) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := TenantFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bookID := r.PathValue("id")
	var body models.UpdateBookBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var book models.Book
	reset, tenantErr := cr.db.UseTenant(context.Background(), tenantID)
	if tenantErr != nil {
		http.Error(w, tenantErr.Error(), http.StatusInternalServerError)
		return
	}
	defer reset()
	if err = cr.db.Model(&book).Where("id = ?", bookID).Updates(models.Book{
		Name: body.Name,
	}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
