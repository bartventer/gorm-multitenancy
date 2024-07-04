package mysql

import (
	"testing"

	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
)

func Test_withMigrationOption(t *testing.T) {
	var newDB = func() *gorm.DB {
		db, err := gorm.Open(tests.DummyDialector{
			TranslatedErr: nil,
		})
		require.NoError(t, err)
		return db
	}

	t.Run("not found", func(t *testing.T) {
		db := newDB()
		_, err := migrationOptionFromDB(db)
		require.ErrorIs(t, err, driver.ErrInvalidMigration)
	})

	t.Run("not of type option", func(t *testing.T) {
		db := newDB()
		err := db.Set(string(migratorKey), "invalid").Error
		require.NoError(t, err)
		_, err = migrationOptionFromDB(db)
		require.ErrorIs(t, err, driver.ErrInvalidMigration)
	})

	t.Run("valid", func(t *testing.T) {
		db := newDB()
		opt := option{"valid"}
		db = withMigrationOption(opt)(db)

		got, err := migrationOptionFromDB(db)
		require.NoError(t, err)
		require.Equal(t, &opt, got)
	})
}
