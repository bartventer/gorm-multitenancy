// Package testutil provides internal testing utilities for the application.
package testutil

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbnamePrefix = "dbname="
)

type DSNOption func(string) string

// WithDBName sets the database name for the connection.
func WithDBName(name string) DSNOption {
	return func(dsn string) string {
		if strings.Contains(dsn, dbnamePrefix) {
			// Replace existing dbname
			start := strings.Index(dsn, dbnamePrefix) + len(dbnamePrefix)
			end := strings.Index(dsn[start:], " ")
			if end == -1 { // If dbname is the last parameter
				end = len(dsn)
			} else {
				end += start
			}
			return dsn[:start] + name + dsn[end:]
		}
		// Append dbname if it doesn't exist
		return fmt.Sprintf("%s %s%s", dsn, dbnamePrefix, name)
	}
}

// GetDSN returns the data source name for the database connection.
func GetDSN(opts ...DSNOption) string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)
	for _, opt := range opts {
		dsn = opt(dsn)
	}
	return dsn
}

// NewTestDB creates a new database connection (for internal use).
func NewTestDB(opts ...DSNOption) *gorm.DB {
	dsn := GetDSN(opts...)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(fmt.Errorf("failed to connect to test database: %w", err))
	}
	return db
}

// DeepEqual compares the expected and actual values using reflect.DeepEqual.
// It returns a boolean indicating whether the values are equal, and a string
// containing an error message if the values are not equal.
func DeepEqual[T any](expected, actual T, message ...interface{}) (bool, string) {
	if !reflect.DeepEqual(expected, actual) {
		// Format the message
		var msg string
		if len(message) > 0 {
			msg = fmt.Sprint(message...)
		} else {
			msg = fmt.Sprintf("Expected %v, got %v", expected, actual)
		}

		return false, fmt.Sprintf("%s: Expected %v, got %v", msg, expected, actual)
	}
	return true, ""
}

// AssertEqual compares two values and logs an error if they are not equal.
func AssertEqual[T any](t *testing.T, expected, actual T, message ...interface{}) bool {
	t.Helper()
	equal, msg := DeepEqual(expected, actual, message...)
	if !equal {
		t.Errorf(msg)
	}
	return equal
}
