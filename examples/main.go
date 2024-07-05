package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"

	"github.com/bartventer/gorm-multitenancy/examples/v8/echoserver"
	"github.com/bartventer/gorm-multitenancy/examples/v8/models"
	"github.com/bartventer/gorm-multitenancy/examples/v8/nethttpserver"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
)

var once sync.Once

type options struct {
	server string // server is the http server. Default is "echo". Options are "echo" and "nethttp".
	driver string // driver is the database driver. Default is "postgres". Options are "postgres" and "mysql".
}

var opts options

func (o *options) validate() error {
	validServers := []string{"echo", "nethttp"}
	validDrivers := []string{"postgres", "mysql"}
	if !slices.Contains(validServers, o.server) {
		return fmt.Errorf("invalid server: %s", o.server)
	}
	if !slices.Contains(validDrivers, o.driver) {
		return fmt.Errorf("invalid driver: %s", o.driver)
	}
	return nil
}

func main() {
	flag.StringVar(&opts.driver, "driver", "postgres", "Specifies the database driver to use. Options: 'postgres', 'mysql'.")
	flag.StringVar(&opts.server, "server", "echo", "Specifies the HTTP server to run and the gorm-multitenancy middleware to use. Options: 'echo', 'nethttp'.")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), `
Examples:

  %s -server=echo -driver=postgres
  %s -server=nethttp -driver=mysql

Note: The server and driver flags are optional. When not specified, the default values are used.
`, os.Args[0], os.Args[0])
		os.Exit(2)
	}
	flag.Parse()
	if err := opts.validate(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		flag.Usage()
	}

	log.SetPrefix("[gorm-multitenancy/examples📦]: ")
	log.SetFlags(log.LstdFlags)
	log.Printf("Starting server with driver: %q, server: %q", opts.driver, opts.server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, cleanup, err := connectDB(ctx, opts.driver)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	err = createExampleData(ctx, db)
	if err != nil {
		panic(err)
	}

	// Start server
	switch opts.server {
	case "nethttp":
		nethttpserver.Start(db)
	case "echo":
		echoserver.Start(db)
	default:
		log.Println("invalid server")
	}

	log.Println("Adios! 🚀")
}

func createExampleData(ctx context.Context, db *multitenancy.DB) (err error) {
	once.Do(func() {
		log.Println("Creating example data...")
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
			log.Printf("Created %d books for tenant: %s", len(books), tenant.SchemaName)
			for _, book := range books {
				log.Printf("Book ID:%d Name:%s TenantSchema:%s", book.ID, book.Name, book.TenantSchema)
			}
		}
	})

	if err != nil {
		log.Printf("Failed to create example data: %v", err)
		return err
	}

	log.Println("OK. Example data created.")
	return nil
}
