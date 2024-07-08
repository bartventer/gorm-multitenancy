package ginserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/models"
	ginmw "github.com/bartventer/gorm-multitenancy/middleware/gin/v8"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/scopes"
	"github.com/gin-gonic/gin"
)

type controller struct {
	db   *multitenancy.DB
	once sync.Once
}

func Start(ctx context.Context, db *multitenancy.DB) error {
	cr := &controller{db: db}
	return cr.start(ctx)
}

func (cr *controller) start(ctx context.Context) (err error) {
	cr.once.Do(func() {
		r := gin.Default()
		r.Use(ginmw.WithTenant(ginmw.WithTenantConfig{
			Skipper: func(c *gin.Context) bool {
				return strings.HasPrefix(c.Request.URL.Path, "/tenants") // skip tenant routes
			},
		}))

		r.POST("/tenants", cr.createTenantHandler)
		r.GET("/tenants/:id", cr.getTenantHandler)
		r.DELETE("/tenants/:id", cr.deleteTenantHandler)
		r.GET("/books", cr.getBooksHandler)
		r.POST("/books", cr.createBookHandler)
		r.DELETE("/books/:id", cr.deleteBookHandler)
		r.PUT("/books/:id", cr.updateBookHandler)

		srv := &http.Server{
			Addr:         ":8080",
			Handler:      r,
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

func TenantFromContext(c *gin.Context) (string, error) {
	tenantID, ok := c.Get(ginmw.TenantKey.String())
	if !ok {
		return "", errors.New("no tenant in context")
	}
	return tenantID.(string), nil
}

func (cr *controller) createTenantHandler(c *gin.Context) {
	var body models.CreateTenantBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	subdomain, subdomainErr := ginmw.ExtractSubdomain(body.DomainURL)
	if subdomainErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": subdomainErr.Error()})
		return
	}
	tenant := &models.Tenant{
		TenantModel: multitenancy.TenantModel{
			DomainURL:  body.DomainURL,
			SchemaName: subdomain,
		},
	}
	if err := cr.db.Create(tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := cr.db.MigrateTenantModels(context.Background(), tenant.SchemaName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := &models.TenantResponse{
		ID:        tenant.ID,
		DomainURL: tenant.DomainURL,
	}
	c.JSON(http.StatusCreated, res)
}

func (cr *controller) getTenantHandler(c *gin.Context) {
	tenantID := c.Param("id")
	tenant := &models.TenantResponse{}
	if err := cr.db.Table(models.TableNameTenant).First(tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tenant)
}

func (cr *controller) deleteTenantHandler(c *gin.Context) {
	tenantID := c.Param("id")
	tenant := &models.Tenant{}
	if err := cr.db.First(tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := cr.db.OffboardTenant(context.Background(), tenant.SchemaName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := cr.db.Delete(&models.Tenant{}, tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (cr *controller) getBooksHandler(c *gin.Context) {
	tenantID, err := TenantFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var books []models.BookResponse
	if err := cr.db.Table(models.TableNameBook).Scopes(scopes.WithTenantSchema(tenantID)).Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, books)
}

func (cr *controller) createBookHandler(c *gin.Context) {
	tenantID, err := TenantFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	book.TenantSchema = tenantID
	reset, tenantErr := cr.db.UseTenant(context.Background(), tenantID)
	if tenantErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tenantErr.Error()})
		return
	}
	defer reset()
	if err := cr.db.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := &models.BookResponse{
		ID:   book.ID,
		Name: book.Name,
	}
	c.JSON(http.StatusCreated, res)
}

func (cr *controller) deleteBookHandler(c *gin.Context) {
	tenantID, err := TenantFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	bookID := c.Param("id")
	var book models.Book
	if err := cr.db.Scopes(scopes.WithTenantSchema(tenantID)).First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := cr.db.Scopes(scopes.WithTenantSchema(tenantID)).Delete(&models.Book{}, bookID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (cr *controller) updateBookHandler(c *gin.Context) {
	tenantID, err := TenantFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	bookID := c.Param("id")
	var body models.UpdateBookBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	book := &models.Book{}
	reset, tenantErr := cr.db.UseTenant(context.Background(), tenantID)
	if tenantErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tenantErr.Error()})
		return
	}
	defer reset()
	if err := cr.db.Model(book).Where("id = ?", bookID).Updates(models.Book{Name: body.Name}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
