package servertest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/initdb"
	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/models"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Harness interface {
	MakeHandler(ctx context.Context, db *multitenancy.DB) (http.Handler, error)
}

func RunConformance(t *testing.T, harness Harness) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, cleanup, err := initdb.Connect(context.Background(), "mysql", func(o *initdb.Options) {
		o.MySQLInitScriptFilePath = filepath.Join("..", "..", "testdata", "init.sql")
	})
	require.NoError(t, err)
	t.Cleanup(cleanup)

	cedoOptions := initdb.CreateExampleDataOptions{
		TenantCount: 2,
		BookCount:   5,
	}
	err = initdb.CreateExampleData(ctx, db, func(cedo *initdb.CreateExampleDataOptions) {
		*cedo = cedoOptions
	})
	require.NoError(t, err)

	handler, err := harness.MakeHandler(ctx, db)
	require.NoError(t, err)

	t.Run("CreateTenant", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/tenants", strings.NewReader(`{"domainUrl": "tenant3.example.com"}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.JSONEq(t, `{"id": 3, "domainUrl": "tenant3.example.com"}`, rr.Body.String())
	})

	t.Run("GetTenant", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/tenants/3", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `{"id": 3, "domainUrl": "tenant3.example.com"}`, rr.Body.String())
	})

	t.Run("DeleteTenant", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/tenants/3", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("GetBooks", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/books", nil)
		require.NoError(t, err)
		req.Host = "tenant1.example.com"

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		tenant := initdb.MakeTenant(1)
		expectedBooks := make([]*models.BookResponse, cedoOptions.BookCount)
		for i := range cedoOptions.BookCount {
			book := initdb.MakeBook(tenant, i+1)
			book.ID = uint(i + 1)
			expectedBooks[i] = &models.BookResponse{
				ID:   book.ID,
				Name: book.Name,
			}
		}

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, toJSON(t, expectedBooks), rr.Body.String())
	})

	t.Run("CreateBook", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/books", strings.NewReader(`{"name": "tenant1 - New Book"}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "tenant1.example.com"

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.JSONEq(t, `{"id": 6, "name": "tenant1 - New Book"}`, rr.Body.String())
	})

	t.Run("DeleteBook", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/books/6", nil)
		require.NoError(t, err)
		req.Host = "tenant1.example.com"

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("UpdateBook", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "/books/2", strings.NewReader(`{"name": "tenant1 - Book 2 - Updated"}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Host = "tenant1.example.com"

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func toJSON(t *testing.T, v interface{}) string {
	t.Helper()
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
