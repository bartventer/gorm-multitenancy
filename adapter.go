package multitenancy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
	"gorm.io/gorm"
)

// Adapater describes the interface for enhancing existing [gorm.DB] instances with additional
// functionalities.
type Adapter interface {
	// AdaptDB takes an existing [gorm.DB] instance and returns a new [DB] instance adapted with
	// additional functionalities.
	//
	// The returned [DB] instance is intended to be used by a single goroutine at a time,
	// ensuring thread safety and avoiding concurrent access issues.
	AdaptDB(ctx context.Context, db *gorm.DB) (*DB, error)

	// OpenDBURL creates a new [DB] instance using the provided URL and returns it.
	// It returns an error if the URL is invalid or if the adapter fails to open the database.
	//
	// The URL must be in a standard url format, as the scheme is used to determine the driver to use.
	OpenDBURL(ctx context.Context, u *url.URL, opts ...gorm.Option) (*DB, error)
}

// driverMux acts as a registry for database driver openers, allowing dynamic driver management.
type driverMux struct {
	mu      sync.RWMutex       // Protects access to the drivers map.
	drivers map[string]Adapter // Maps driver names to their respective openers.
}

// Register adds a new adapter to the registry under the specified driver name.
// It panics if a Adapter for the given driver name is already registered.
func (mux *driverMux) Register(driver string, adapter Adapter) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	if mux.drivers == nil {
		mux.drivers = make(map[string]Adapter)
	}
	if _, exists := mux.drivers[driver]; exists {
		panic(gmterrors.New(errors.New("driver already registered: " + driver)))
	}
	mux.drivers[driver] = adapter
}

// AdaptDB creates a new [DB] instance using the provided db instance and driver name.
// It returns an error if no adapter is registered for the given driver name.
func (mux *driverMux) AdaptDB(ctx context.Context, db *gorm.DB) (*DB, error) {
	driverName := db.Name()
	mux.mu.RLock()
	adapter, ok := mux.drivers[driverName]
	mux.mu.RUnlock()
	if !ok {
		return nil, gmterrors.New(errors.New("no registered adapter for driver: " + driverName))
	}
	return adapter.AdaptDB(ctx, db)
}

// OpenDB creates a new [DB] instance using the provided URL string and returns it.
// It returns an error if the URL is invalid or if the adapter fails to open the database.
func (mux *driverMux) OpenDB(ctx context.Context, urlstr string, opts ...gorm.Option) (*DB, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, gmterrors.New(fmt.Errorf("failed to parse URL: %w", err))
	}
	driverName := u.Scheme
	mux.mu.RLock()
	adapter, ok := mux.drivers[driverName]
	mux.mu.RUnlock()
	if !ok {
		return nil, gmterrors.New(errors.New("no registered adapter for driver: " + driverName))
	}
	return adapter.OpenDBURL(ctx, u, opts...)
}

var defaultDriverMux = new(driverMux)

// Register adds a new [Adapter] to the default registry under the specified driver name.
// It panics if an [Adapter] for the given driver name is already registered.
func Register(name string, adapter Adapter) {
	defaultDriverMux.Register(name, adapter)
}

// Open is a drop-in replacement for [gorm.Open]. It returns a new [DB] instance using
// the provided dialector and options.
//
// MySQL:
//
//	import (
//			"github.com/bartventer/gorm-multitenancy/mysql/v8"
//			multitenancy "github.com/bartventer/gorm-multitenancy/v8"
//	 )
//
//	 func main() {
//			dsn := "user:password@tcp(localhost:3306)/dbname?parseTime=True"
//			db, err := multitenancy.Open(mysql.Open(dsn))
//			if err != nil {...}
//	 }
//
// PostgreSQL:
//
//	 import (
//	 	"github.com/bartventer/gorm-multitenancy/postgres/v8"
//	 	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
//	 )
//
//	func main() {
//		dsn := "postgres://user:password@localhost:5432/dbname?sslmode=disable"
//		db, err := multitenancy.Open(postgres.Open(dsn))
//		if err != nil {...}
//	}
func Open(dialector gorm.Dialector, opts ...gorm.Option) (*DB, error) {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, gmterrors.New(fmt.Errorf("failed to open gorm database: %w", err))
	}
	return defaultDriverMux.AdaptDB(context.TODO(), db)
}

// OpenDB creates a new [DB] instance using the provided URL string and returns it.
//
// The URL string must be in a standard url format, as the scheme is used to determine
// the driver to use. The following are examples of valid URL strings for MySQL and
// PostgreSQL:
//
//	"mysql://user:password@tcp(localhost:3306)/dbname"
//	"postgres://user:password@localhost:5432/dbname"
//
// MySQL:
//
//	import (
//		_ "github.com/bartventer/gorm-multitenancy/mysql/v8"
//		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
//	)
//
//	func main() {
//		url := "mysql://user:password@tcp(localhost:3306)/dbname"
//		db, err := multitenancy.OpenDB(context.Background(), url)
//		if err != nil {...}
//	}
//
// PostgreSQL:
//
//	import (
//		_ "github.com/bartventer/gorm-multitenancy/postgres/v8"
//		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
//	)
//
//	func main() {
//		url := "postgres://user:password@localhost:5432/dbname"
//		db, err := multitenancy.OpenDB(context.Background(), url)
//		if err != nil {...}
//	}
func OpenDB(ctx context.Context, urlstr string) (*DB, error) {
	return defaultDriverMux.OpenDB(ctx, urlstr)
}
