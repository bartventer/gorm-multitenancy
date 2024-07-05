// Package testutil provides internal testing utilities for the application.
package testutil

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

type dbContainer struct {
	opts          []gorm.Option
	dialectOpener func(dsn string) gorm.Dialector
}

func (c *dbContainer) NewDB(t testing.TB, ctx context.Context) *gorm.DB {
	t.Helper()
	dbName := "tenants"
	dbUser := "tenants"
	dbPassword := "tenants"
	dbPort := "5432/tcp"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{dbPort},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		Tmpfs:      map[string]string{"/var/lib/postgres": "rw"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Memory = 512 * 1024 * 1024 // 512MB
			hc.NanoCPUs = 500000000       // 0.5 CPU
		},
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoErrorf(t, err, "Failed to start postgres container: %v", err)
	t.Cleanup(func() {
		cleanupErr := postgresContainer.Terminate(ctx)
		require.NoErrorf(t, cleanupErr, "Failed to terminate Postgres container: %v", cleanupErr)
	})

	host, err := postgresContainer.Host(ctx)
	require.NoErrorf(t, err, "Failed to get postgres container host: %v", err)
	natPort, err := postgresContainer.MappedPort(ctx, nat.Port(dbPort))
	require.NoErrorf(t, err, "Failed to get postgres container port: %v", err)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, natPort.Port(), dbUser, dbPassword, dbName)
	db, err := gorm.Open(c.dialectOpener(dsn), c.opts...)
	require.NoErrorf(t, err, "Failed to connect to database: %v", err)

	summary := makeContainerSummary(t, postgresContainer)
	log.Println(summary)

	return db
}

// NewDBWithOptions creates a new database instance for testing with the provided options.
func NewDBWithOptions(t testing.TB, ctx context.Context, opener func(dsn string) gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	t.Helper()
	c := dbContainer{
		dialectOpener: opener,
		opts:          opts,
	}
	return c.NewDB(t, ctx)
}

// makeContainerSummary generates a summary of the container for debugging purposes.
func makeContainerSummary(t testing.TB, container testcontainers.Container) string {
	t.Helper()
	containerJSON, err := container.Inspect(context.Background())
	require.NoErrorf(t, err, "Failed to inspect container: %v", err)
	return fmt.Sprintf(`
PostgreSQL:
	Test: %s
	Status: %s
    Ports: %+v
    Env: %q

`,
		t.Name(),
		containerJSON.State.Status,
		containerJSON.NetworkSettings.Ports,
		containerJSON.Config.Env,
	)
}
