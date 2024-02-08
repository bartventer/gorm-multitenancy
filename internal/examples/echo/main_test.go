package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	echomw "github.com/bartventer/gorm-multitenancy/v3/middleware/echo"
	"github.com/labstack/echo/v4"
)

func TestTenantFromContext(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid: Tenant exists in context",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "Invalid: Tenant does not exist in context",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TenantFromContext(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("TenantFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TenantFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_createTenantHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid request",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					body := &CreateTenantBody{
						DomainURL: fmt.Sprintf("tenant%d.example.com", time.Now().UnixNano()),
					}
					bodyBytes, _ := json.Marshal(body)
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			wantErr: false,
		},
		{
			name: "Invalid request",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodPost, "/", nil)
					rec := httptest.NewRecorder()
					return e.NewContext(req, rec)
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createTenantHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("createTenantHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getTenantHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Tenant exists",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/tenants/1", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("tenants/:id")
					c.SetParamNames("id")
					c.SetParamValues("1")
					return c
				}(),
			},
			wantErr: false,
		},
		{
			name: "Tenant does not exist",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/tenants/999", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("tenants/:id")
					c.SetParamNames("id")
					c.SetParamValues("999")
					return c
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getTenantHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("getTenantHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteTenantHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Tenant exists",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodDelete, "/tenants/1", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("tenants/:id")
					c.SetParamNames("id")
					c.SetParamValues("1")
					return c
				}(),
			},
			wantErr: false,
		},
		{
			name: "Tenant does not exist",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodDelete, "/tenants/999", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("tenants/:id")
					c.SetParamNames("id")
					c.SetParamValues("999")
					return c
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteTenantHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("deleteTenantHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_getBooksHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Tenant exists and has books",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/books", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: false,
		},
		{
			name: "Tenant does not exist",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/books", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.Set(echomw.TenantKey.String(), "tenant999")
					return c
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getBooksHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("getBooksHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_createBookHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid book creation",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					bookJSON := `{"name":"Test Book"}`
					req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: false,
		},
		{
			name: "Invalid book creation",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					bookJSON := `{"name":""}` // Invalid because name is empty
					req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createBookHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("createBookHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteBookHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid book deletion",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("books/:id")
					c.SetParamNames("id")
					c.SetParamValues("1")
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: false,
		},
		{
			name: "Invalid book deletion",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodDelete, "/books/999", nil)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("books/:id")
					c.SetParamNames("id")
					c.SetParamValues("999")
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteBookHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("deleteBookHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_updateBookHandler(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid book update",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					bookJSON := `{"name":"Updated Test Book"}`
					req := httptest.NewRequest(http.MethodPut, "/books/2", strings.NewReader(bookJSON))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("books/:id")
					c.SetParamNames("id")
					c.SetParamValues("1")
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: false,
		},
		{
			name: "Invalid book update",
			args: args{
				c: func() echo.Context {
					e := echo.New()
					bookJSON := `{"name":""}` // Invalid because name is empty
					req := httptest.NewRequest(http.MethodPut, "/books/2", strings.NewReader(bookJSON))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					c := e.NewContext(req, rec)
					c.SetPath("books/:id")
					c.SetParamNames("id")
					c.SetParamValues("1")
					c.Set(echomw.TenantKey.String(), "tenant1")
					return c
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateBookHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("updateBookHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
