package initdb

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/models"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/fatih/color"
)

var once sync.Once

func CreateExampleData(ctx context.Context, db *multitenancy.DB) (err error) {
	once.Do(func() {
		color.Set(color.FgYellow, color.Bold)
		defer color.Unset()
		log.Println("Creating example data...")
		log.Println("This may take a few seconds...")
		if err = db.RegisterModels(ctx, &models.Tenant{}, &models.Book{}); err != nil {
			return
		}

		if err = db.MigrateSharedModels(ctx); err != nil {
			return
		}

		tenants := []*models.Tenant{
			{
				TenantModel: multitenancy.TenantModel{
					DomainURL:  "tenant1.example.com",
					SchemaName: "tenant1",
				},
			},
			{
				TenantModel: multitenancy.TenantModel{
					DomainURL:  "tenant2.example.com",
					SchemaName: "tenant2",
				},
			},
		}
		if err = db.Create(&tenants).Error; err != nil {
			return
		}
		color.Set(color.FgYellow)
		log.Printf("Created %d tenants", len(tenants))
		for _, tenant := range tenants {
			log.Printf("Tenant ID:%d", tenant.ID)
			log.Printf("%#v", tenant.TenantModel)
		}

		var makeBooks = func(tenant *models.Tenant) []*models.Book {
			var books []*models.Book
			for i := 1; i <= 5; i++ {
				books = append(books, &models.Book{
					Tenant: *tenant,
					Name:   fmt.Sprintf("Book %d", i),
				})
			}
			return books
		}

		for _, tenant := range tenants {
			if err = db.MigrateTenantModels(ctx, tenant.SchemaName); err != nil {
				return
			}
			// Create tenant specific data
			var reset func() error
			reset, err = db.UseTenant(ctx, tenant.SchemaName)
			if err != nil {
				return
			}
			defer reset()
			books := makeBooks(tenant)
			if err = db.Create(books).Error; err != nil {
				return
			}
			color.Set(color.FgYellow)
			log.Printf("Created %d books for tenant: %q", len(books), tenant.SchemaName)
			for _, book := range books {
				log.Printf("Book ID:%d Name:%q TenantSchema:%q", book.ID, book.Name, book.TenantSchema)
			}
		}
		color.Set(color.FgGreen, color.Bold)
		log.Println("OK. Example data created.")
	})

	if err != nil {
		log.Printf("Failed to create example data: %v", err)
		return err
	}

	return nil
}
