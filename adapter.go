package multitenancy

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
	"gorm.io/gorm"
)

type (

	// Adapter defines an interface for enhancing [gorm.DB] instances with additional functionalities.
	Adapter interface {
		// AdaptDB enhances an existing [gorm.DB] instance with additional functionalities and returns
		// a new [DB] instance. The returned DB instance should be used by a single goroutine at a time
		// to ensure thread safety and prevent concurrent access issues.
		AdaptDB(ctx context.Context, db *gorm.DB) (*DB, error)

		// OpenDBURL creates and returns a new [DB] instance using the provided URL. It returns an error
		// if the URL is invalid or the adapter fails to open the database. The URL must follow a standard
		// format, using the scheme to determine the driver.
		OpenDBURL(ctx context.Context, u *driver.URL, opts ...gorm.Option) (*DB, error)
	}

	// adapterMux is a multiplexer that holds a map of driver names to their respective adapters.
	adapterMux struct {
		mu      sync.RWMutex       // Protects access to the drivers map.
		drivers map[string]Adapter // Maps driver names to their respective openers.
	}
)

// Register adds a new adapter to the registry under the specified driver name.
// It panics if a Adapter for the given driver name is already registered.
func (mux *adapterMux) Register(driver string, adapter Adapter) {
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
func (mux *adapterMux) AdaptDB(ctx context.Context, db *gorm.DB) (*DB, error) {
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
func (mux *adapterMux) OpenDB(ctx context.Context, urlstr string, opts ...gorm.Option) (*DB, error) {
	u, err := driver.ParseURL(urlstr)
	if err != nil {
		return nil, gmterrors.New(err)
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

var defaultDriverMux = new(adapterMux)

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

// OpenDB creates a new DB instance using the provided URL string and returns it.
// The URL string must be in a standard URL format, as the scheme is used to determine
// the driver to use. Refer to the driver-specific documentation for more information
// on the URL format.
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
func OpenDB(ctx context.Context, urlstr string, opts ...gorm.Option) (*DB, error) {
	return defaultDriverMux.OpenDB(ctx, urlstr, opts...)
}
