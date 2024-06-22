package postgres

import "fmt"

type contextKey struct {
	name string
}

func (c contextKey) String() string {
	return fmt.Sprintf("gorm-multitenancy/postgres/%s", c.name)
}

var (
	// MigrationOptions is the context key for the migration options.
	MigrationOptions = &contextKey{"migration_options"}
)
