package initdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bartventer/gorm-multitenancy/mysql/v8"
	"github.com/bartventer/gorm-multitenancy/postgres/v8"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

func Connect(ctx context.Context, driver string) (db *multitenancy.DB, cleanup func(), err error) {
	var config = struct {
		User, Password, Name, Port string
	}{
		User:     "gmt",
		Password: "gmt",
		Name:     "gmt",
	}
	var container testcontainers.Container
	switch driver {
	case "postgres":
		config.Port = "5432/tcp"
		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:16-alpine",
				ExposedPorts: []string{config.Port},
				Env: map[string]string{
					"POSTGRES_USER":     config.User,
					"POSTGRES_PASSWORD": config.Password,
					"POSTGRES_DB":       config.Name,
				},
				WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
				Tmpfs:      map[string]string{"/var/lib/postgres": "rw"},
			},
			Started: true,
		})
	case "mysql":
		config.Port = "3306/tcp"
		config.Name = "public"
		initScriptPath := filepath.Join("testdata", "init.sql")
		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "mysql:8.0",
				ExposedPorts: []string{config.Port, "33060/tcp"},
				Env: map[string]string{
					"MYSQL_USER":                 config.User,
					"MYSQL_PASSWORD":             config.Password,
					"MYSQL_ROOT_PASSWORD":        config.Password,
					"MYSQL_DATABASE":             config.Name,
					"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
				},
				WaitingFor: wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(6 * time.Minute),
				Tmpfs:      map[string]string{"/var/lib/mysql": "rw"},
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      initScriptPath,
						Reader:            nil,
						ContainerFilePath: filepath.Join("/docker-entrypoint-initdb.d", filepath.Base(initScriptPath)),
						FileMode:          int64(os.ModePerm),
					},
				},
			},
			Started: true,
		})
	}
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() {
		if terminateErr := container.Terminate(ctx); terminateErr != nil {
			log.Println("Failed to terminate container:", terminateErr)
		}
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, cleanup, err
	}
	natPort, err := container.MappedPort(ctx, nat.Port(config.Port))
	if err != nil {
		return nil, cleanup, err
	}

	var dsn string
	switch driver {
	case "postgres":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			config.User, config.Password, host, natPort.Port(), config.Name)
		db, err = multitenancy.Open(postgres.Open(dsn), &gorm.Config{})
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			config.User, config.Password, host, natPort.Port(), config.Name)
		db, err = multitenancy.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		err = errors.New("invalid driver")
	}
	if err != nil {
		log.Println("Failed to connect to database:", err)
		return nil, cleanup, err
	} else {
		log.Println("Connected to database.")
		log.Printf("DSN: %q", dsn)
	}
	return db, cleanup, nil
}
