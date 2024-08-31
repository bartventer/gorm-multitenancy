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

type CreateExampleDataOptions struct {
	TenantCount int
	BookCount   int
}

func MakeTenant(id int) *models.Tenant {
	return &models.Tenant{
		TenantModel: multitenancy.TenantModel{
			DomainURL:  fmt.Sprintf("tenant%d.example.com", id),
			SchemaName: fmt.Sprintf("tenant%d", id),
		},
	}
}

func MakeBook(tenant *models.Tenant, id int) *models.Book {
	return &models.Book{
		Tenant: *tenant,
		Name:   fmt.Sprintf("Book %d", id),
	}
}

type CreateExampleDataOption func(*CreateExampleDataOptions)

func CreateExampleData(ctx context.Context, db *multitenancy.DB, opts ...CreateExampleDataOption) (err error) {
	once.Do(func() {

		options := CreateExampleDataOptions{
			TenantCount: 2,
			BookCount:   5,
		}
		for _, opt := range opts {
			opt(&options)
		}

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

		tenants := make([]*models.Tenant, options.TenantCount)
		for i := range tenants {
			tenants[i] = MakeTenant(i + 1)
		}
		if err = db.Create(&tenants).Error; err != nil {
			return
		}
		color.Set(color.FgYellow)
		log.Printf("Created %d tenants", len(tenants))
		for _, tenant := range tenants {
			log.Printf("Tenant ID:%d", tenant.ID)
			log.Printf("\t%#v", tenant.TenantModel)
		}

		var makeBooks = func(tenant *models.Tenant) []*models.Book {
			books := make([]*models.Book, options.BookCount)
			for i := range books {
				books[i] = MakeBook(tenant, i+1)
			}
			return books
		}

		for _, tenant := range tenants {
			if err = db.MigrateTenantModels(ctx, tenant.SchemaName); err != nil {
				return
			}
			var reset func() error
			reset, err = db.UseTenant(ctx, tenant.SchemaName)
			if err != nil {
				return
			}
			books := makeBooks(tenant)
			if err = db.Create(books).Error; err != nil {
				reset()
				return
			}
			color.Set(color.FgYellow)
			log.Printf("Created %d books for tenant: %q", len(books), tenant.SchemaName)
			for _, book := range books {
				log.Printf("\tBook ID:%d Name:%q TenantSchema:%q", book.ID, book.Name, book.TenantSchema)
			}

			if err = reset(); err != nil {
				return
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
