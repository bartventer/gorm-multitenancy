// Package testutil provides internal testing utilities for the application.
package testutil

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GetDSN returns the data source name for the database connection.
func GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PASSWORD"),
	)
}

// NewTestDB creates a new database connection (for internal use).
func NewTestDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(GetDSN()), &gorm.Config{PrepareStmt: true})
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to test database"))
	}
	return db
}
