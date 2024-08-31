package irisserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/models"
	irismiddleware "github.com/bartventer/gorm-multitenancy/middleware/iris/v8"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/scopes"
	"github.com/kataras/iris/v12"
)

type controller struct {
	db   *multitenancy.DB
	once sync.Once
}

func Start(ctx context.Context, db *multitenancy.DB) error {
	cr := &controller{db: db}
	return cr.start(ctx)
}

func (c *controller) init(app *iris.Application) {
	app.Use(irismiddleware.WithTenant(irismiddleware.WithTenantConfig{
		Skipper: func(ctx iris.Context) bool {
			return strings.HasPrefix(ctx.Request().URL.Path, "/tenants") // skip tenant routes
		},
	}))

	app.Post("/tenants", c.createTenantHandler)
	app.Get("/tenants/{id:string}", c.getTenantHandler)
	app.Delete("/tenants/{id:string}", c.deleteTenantHandler)
	app.Get("/books", c.getBooksHandler)
	app.Post("/books", c.createBookHandler)
	app.Delete("/books/{id:string}", c.deleteBookHandler)
	app.Put("/books/{id:string}", c.updateBookHandler)
}

func (cr *controller) start(ctx context.Context) (err error) {
	cr.once.Do(func() {
		app := iris.New()
		cr.init(app)

		srv := &http.Server{
			Addr:         ":8080",
			Handler:      app,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go func() {
			if serveErr := srv.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
				log.Printf("listen: %s\n", serveErr)
				err = serveErr
			}
		}()

		<-ctx.Done()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if shutdownErr := srv.Shutdown(ctxShutdown); shutdownErr != nil {
			log.Printf("Server forced to shutdown: %v", shutdownErr)
			if err == nil {
				err = shutdownErr
			}
		}

		log.Println("Server exiting")
	})
	return err
}

func TenantFromContext(ctx iris.Context) (string, error) {
	tenantID := ctx.Values().GetString(irismiddleware.TenantKey.String())
	if tenantID == "" {
		return "", errors.New("no tenant in context")
	}
	return tenantID, nil
}

func (cr *controller) createTenantHandler(ctx iris.Context) {
	var body models.CreateTenantBody
	if err := ctx.ReadJSON(&body); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	subdomain, subdomainErr := irismiddleware.ExtractSubdomain(body.DomainURL)
	if subdomainErr != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": subdomainErr.Error()})
		return
	}
	tenant := &models.Tenant{
		TenantModel: multitenancy.TenantModel{
			DomainURL:  body.DomainURL,
			SchemaName: subdomain,
		},
	}
	if err := cr.db.Create(tenant).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	if err := cr.db.MigrateTenantModels(context.Background(), tenant.SchemaName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	res := &models.TenantResponse{
		ID:        tenant.ID,
		DomainURL: tenant.DomainURL,
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(res)
}

func (cr *controller) getTenantHandler(ctx iris.Context) {
	tenantID := ctx.Params().Get("id")
	tenant := &models.TenantResponse{}
	if err := cr.db.Table(models.TableNameTenant).First(tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	ctx.JSON(tenant)
}

func (cr *controller) deleteTenantHandler(ctx iris.Context) {
	tenantID := ctx.Params().Get("id")
	tenant := &models.Tenant{}
	if err := cr.db.First(tenant, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	if err := cr.db.OffboardTenant(context.Background(), tenant.SchemaName); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	if err := cr.db.Delete(&models.Tenant{}, tenantID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

func (cr *controller) getBooksHandler(ctx iris.Context) {
	tenantID, err := TenantFromContext(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	var books []models.BookResponse
	if err := cr.db.Table(models.TableNameBook).Scopes(scopes.WithTenantSchema(tenantID)).Find(&books).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	ctx.JSON(books)
}

func (cr *controller) createBookHandler(ctx iris.Context) {
	tenantID, err := TenantFromContext(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	var book models.Book
	if err := ctx.ReadJSON(&book); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	book.TenantSchema = tenantID
	reset, tenantErr := cr.db.UseTenant(context.Background(), tenantID)
	if tenantErr != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": tenantErr.Error()})
		return
	}
	defer reset()
	if err := cr.db.Create(&book).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}

	res := &models.BookResponse{
		ID:   book.ID,
		Name: book.Name,
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(res)
}

func (cr *controller) deleteBookHandler(ctx iris.Context) {
	tenantID, err := TenantFromContext(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	bookID := ctx.Params().Get("id")
	var book models.Book
	if err := cr.db.Scopes(scopes.WithTenantSchema(tenantID)).First(&book, bookID).Error; err != nil {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	if err := cr.db.Scopes(scopes.WithTenantSchema(tenantID)).Delete(&models.Book{}, bookID).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

func (cr *controller) updateBookHandler(ctx iris.Context) {
	tenantID, err := TenantFromContext(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	bookID := ctx.Params().Get("id")
	var body models.UpdateBookBody
	if err := ctx.ReadJSON(&body); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	if body.Name == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "name is required"})
		return
	}
	book := &models.Book{}
	reset, tenantErr := cr.db.UseTenant(context.Background(), tenantID)
	if tenantErr != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": tenantErr.Error()})
		return
	}
	defer reset()
	if err := cr.db.Model(book).Where("id = ?", bookID).Updates(models.Book{Name: body.Name}).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
}
