package internal

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewTestDB creates a new database connection (for internal use)
func NewTestDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)), &gorm.Config{PrepareStmt: true})
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to test database"))
	}
	return db
}
