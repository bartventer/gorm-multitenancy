package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bartventer/gorm-multitenancy/v3/tenantcontext"
	"github.com/go-chi/chi/v5"
)

func Test_createTenantHandler(t *testing.T) {
	tests := []struct {
		name    string
		body    *CreateTenantBody
		wantErr bool
	}{
		{
			name: "Valid request",
			body: &CreateTenantBody{
				DomainURL: fmt.Sprintf("tenant%d.example.com", time.Now().UnixNano()),
			},
			wantErr: false,
		},
		{
			name:    "Invalid request",
			body:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Register the handler with the router
			r.Post("/", createTenantHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusCreated {
				t.Errorf("createTenantHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusCreated {
				t.Errorf("createTenantHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}
func Test_getTenantHandler(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Tenant exists",
			url:     "/tenants/1",
			wantErr: false,
		},
		{
			name:    "Tenant does not exist",
			url:     "/tenants/999",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Register the handler with the router
			r.Get("/tenants/{id}", getTenantHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusOK {
				t.Errorf("getTenantHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusOK {
				t.Errorf("getTenantHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}

func Test_deleteTenantHandler(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Tenant exists",
			url:     "/tenants/1",
			wantErr: false,
		},
		{
			name:    "Tenant does not exist",
			url:     "/tenants/999",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Register the handler with the router
			r.Delete("/tenants/{id}", deleteTenantHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusNoContent {
				t.Errorf("deleteTenantHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusNoContent {
				t.Errorf("deleteTenantHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}
func Test_getBooksHandler(t *testing.T) {
	type args struct {
		tenant string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Tenant exists and has books",
			args: args{
				tenant: "tenant1",
			},
			wantErr: false,
		},
		{
			name: "Tenant does not exist",
			args: args{
				tenant: "tenant999",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/books", nil)
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Create a middleware that sets the tenant in the context
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := context.WithValue(r.Context(), tenantcontext.TenantKey, tt.args.tenant)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			// Register the handler with the router
			r.Get("/books", getBooksHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusOK {
				t.Errorf("getBooksHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusOK {
				t.Errorf("getBooksHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}

func Test_createBookHandler(t *testing.T) {
	type args struct {
		tenant   string
		bookJSON string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid book creation",
			args: args{
				tenant:   "tenant1",
				bookJSON: `{"name":"Test Book"}`,
			},
			wantErr: false,
		},
		{
			name: "Invalid book creation",
			args: args{
				tenant:   "tenant1",
				bookJSON: `{"name":""}`, // Invalid because name is empty
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(tt.args.bookJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Create a middleware that sets the tenant in the context
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := context.WithValue(r.Context(), tenantcontext.TenantKey, tt.args.tenant)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			// Register the handler with the router
			r.Post("/books", createBookHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusCreated {
				t.Errorf("createBookHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusCreated {
				t.Errorf("createBookHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}

func Test_deleteBookHandler(t *testing.T) {
	type args struct {
		tenant string
		bookID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid book deletion",
			args: args{
				tenant: "tenant1",
				bookID: "1",
			},
			wantErr: false,
		},
		{
			name: "Invalid book deletion",
			args: args{
				tenant: "tenant1",
				bookID: "999",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/books/"+tt.args.bookID, nil)
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Create a middleware that sets the tenant in the context
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := context.WithValue(r.Context(), tenantcontext.TenantKey, tt.args.tenant)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			// Register the handler with the router
			r.Delete("/books/{id}", deleteBookHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusNoContent {
				t.Errorf("deleteBookHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusNoContent {
				t.Errorf("deleteBookHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}

func Test_updateBookHandler(t *testing.T) {
	type args struct {
		tenant string
		bookID string
		body   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid book update",
			args: args{
				tenant: "tenant1",
				bookID: "1",
				body:   `{"name":"Updated Test Book"}`,
			},
			wantErr: false,
		},
		{
			name: "Invalid book update",
			args: args{
				tenant: "tenant1",
				bookID: "1",
				body:   `{"name":""}`, // Invalid because name is empty
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/books/"+tt.args.bookID, strings.NewReader(tt.args.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Create a new Chi router
			r := chi.NewRouter()

			// Create a middleware that sets the tenant in the context
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := context.WithValue(r.Context(), tenantcontext.TenantKey, tt.args.tenant)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})

			// Register the handler with the router
			r.Put("/books/{id}", updateBookHandler)

			// Pass the request to the router
			r.ServeHTTP(rec, req)

			if tt.wantErr && rec.Code == http.StatusOK {
				t.Errorf("updateBookHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
			if !tt.wantErr && rec.Code != http.StatusOK {
				t.Errorf("updateBookHandler() error = %v, wantErr %v", rec.Code, tt.wantErr)
			}
		})
	}
}
