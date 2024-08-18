package driver_test

import (
	"fmt"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
)

func ExampleParseURL() {
	dsn := "mysql://user:password@tcp(localhost:3306)/dbname"
	u, err := driver.ParseURL(dsn)
	if err != nil {
		panic(err)
	}

	fmt.Println("Scheme:", u.Scheme)
	fmt.Println("Host:", u.Host)
	fmt.Println("Path:", u.Path)
	fmt.Println("User:", u.User.String())
	fmt.Println("Raw URL:", u.Raw())
	fmt.Println("Sanitized URL:", u.String())

	// Output:
	// Scheme: mysql
	// Host: localhost:3306
	// Path: /dbname
	// User: user:password
	// Raw URL: mysql://user:password@tcp(localhost:3306)/dbname
	// Sanitized URL: mysql://user:password@localhost:3306/dbname
}
