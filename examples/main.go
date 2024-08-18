package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/echoserver"
	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/ginserver"
	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/initdb"
	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/nethttpserver"
	"github.com/fatih/color"
)

type options struct {
	server string // server is the http server. Default is "echo". Options are "echo" and "nethttp".
	driver string // driver is the database driver. Default is "postgres". Options are "postgres" and "mysql".
}

var opts options

func (o *options) validate() error {
	validServers := []string{"echo", "gin", "nethttp"}
	if !slices.Contains(validServers, o.server) {
		return fmt.Errorf("invalid server: %s", o.server)
	}
	validDrivers := []string{"postgres", "mysql"}
	if !slices.Contains(validDrivers, o.driver) {
		return fmt.Errorf("invalid driver: %s", o.driver)
	}
	return nil
}

func main() {
	color.Cyan(`
	Welcome to the GORM Multitenancy Example Application! 🚀
	We're glad you're here. Let's explore multitenancy together.
	`)

	color.Yellow(`
	Resources:
	- API Usage: https://github.com/bartventer/gorm-multitenancy/tree/master/examples/USAGE.md
	- Documentation & Guides: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v8
	`)

	flag.StringVar(&opts.driver, "driver", "postgres", "Specifies the database driver to use. Options: 'postgres', 'mysql'.")
	flag.StringVar(&opts.server, "server", "echo", "Specifies the HTTP server to run and the gorm-multitenancy middleware to use. Options: 'echo', 'gin', 'nethttp'.")
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
		color.Red("error: %v\n", err)
		flag.Usage()
	}

	log.SetPrefix("[gorm-multitenancy/examples📦]: ")
	log.SetFlags(log.LstdFlags)
	color.Magenta(`
	Starting server with the following options:
	- Server: %s
	- Driver: %s
	`, opts.server, opts.driver)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down server... 🛑")
		cancel()
	}()

	db, cleanup, err := initdb.Connect(ctx, opts.driver)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer cleanup()
	err = initdb.CreateExampleData(ctx, db)
	if err != nil {
		log.Fatalf("Failed to create example data: %v\n", err)
	}

	switch opts.server {
	case "echo":
		err = echoserver.Start(ctx, db)
	case "gin":
		err = ginserver.Start(ctx, db)
	case "nethttp":
		err = nethttpserver.Start(ctx, db)
	default:
		err = fmt.Errorf("invalid server: %s", opts.server)
	}

	if err != nil {
		log.Printf("Failed to start server: %v\n", err)
		cancel()
	}

	<-ctx.Done()

	color.Magenta(`
	Thank you for using the GORM Multitenancy Example Application! 🙏
	We hope you found it informative and helpful.
	If you have any questions or feedback, please let us know.
	
	For more information, visit https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v8

	Until next time! Happy coding! 🚀
	`)

	time.Sleep(2 * time.Second)
	os.Exit(0)
}
