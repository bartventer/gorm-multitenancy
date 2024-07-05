package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/echoserver"
	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/initdb"
	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/nethttpserver"
)

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
	db, cleanup, err := initdb.Connect(ctx, opts.driver)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	err = initdb.CreateExampleData(ctx, db)
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
