// Package testutil provides utility functions for testing.
package testutil

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

const initSQL = `
CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON *.* TO '%s'@'%%' WITH GRANT OPTION;
FLUSH PRIVILEGES;
`

type dbContainer struct {
	opts          []gorm.Option
	dialectOpener func(dsn string) gorm.Dialector
}

// generateInitSQL creates an init.sql file with the necessary SQL commands to create a user and grant privileges.
func (c *dbContainer) generateInitSQL(t testing.TB, dbUser string) (file string) {
	t.Helper()
	sqlContent := fmt.Sprintf(initSQL, dbUser, dbUser)
	dir := t.TempDir()
	file = fmt.Sprintf("%s/init.sql", dir)
	err := os.WriteFile(file, []byte(sqlContent), 0600)
	require.NoErrorf(t, err, "Failed to write init.sql: %v", err)
	return file
}

func (c *dbContainer) newDB(t testing.TB, ctx context.Context) *gorm.DB {
	t.Helper()
	dbName := "public"
	dbUser := "tenants"
	dbPassword := "tenants"

	initScript := c.generateInitSQL(t, dbUser)

	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_DATABASE":             dbName,
			"MYSQL_USER":                 dbUser,
			"MYSQL_PASSWORD":             dbPassword,
			"MYSQL_ROOT_PASSWORD":        dbPassword,
			"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
		},
		WaitingFor: wait.
			ForLog("port: 3306  MySQL Community Server").
			WithStartupTimeout(15 * time.Minute),
		Tmpfs: map[string]string{"/var/lib/mysql": "rw"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      initScript,
				Reader:            nil,
				ContainerFilePath: filepath.Join("/docker-entrypoint-initdb.d", filepath.Base(initScript)),
				FileMode:          int64(os.ModePerm),
			},
		},
	}

	mysqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoErrorf(t, err, "Failed to start MySQL container: %v", err)
	t.Cleanup(func() {
		cleanupErr := mysqlContainer.Terminate(ctx)
		require.NoErrorf(t, cleanupErr, "Failed to terminate MySQL container: %v", cleanupErr)
	})

	host, err := mysqlContainer.Host(ctx)
	require.NoErrorf(t, err, "Failed to get MySQL container host: %v", err)
	natPort, err := mysqlContainer.MappedPort(ctx, "3306/tcp")
	require.NoErrorf(t, err, "Failed to get MySQL container port: %v", err)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser,
		dbPassword,
		host,
		natPort.Port(),
		dbName,
	)
	db, err := gorm.Open(c.dialectOpener(dsn), c.opts...)
	require.NoErrorf(t, err, "Failed to connect to database: %v", err)

	summary := makeContainerSummary(t, mysqlContainer, dsn)
	log.Println(summary)

	return db
}

// NewDB creates a new database instance for testing with the provided options.
func NewDB(t testing.TB, ctx context.Context, opener func(dsn string) gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	t.Helper()
	c := dbContainer{
		dialectOpener: opener,
		opts:          opts,
	}
	return c.newDB(t, ctx)
}

// makeContainerSummary generates a summary of the container for debugging purposes.
func makeContainerSummary(t testing.TB, container testcontainers.Container, dsn string) string {
	t.Helper()
	containerJSON, err := container.Inspect(context.Background())
	require.NoErrorf(t, err, "Failed to inspect container: %v", err)
	u, err := driver.ParseURL(dsn)
	require.NoErrorf(t, err, "Failed to parse DSN: %v", err)
	return fmt.Sprintf(`
MySQL:
	Test: %s
    Status: %s
	DSN: %s
    Ports: %+v
    Env: %q

`,
		t.Name(),
		containerJSON.State.Status,
		"mysql://"+strings.TrimPrefix(u.String(), u.Scheme+"://"),
		containerJSON.NetworkSettings.Ports,
		containerJSON.Config.Env,
	)
}
