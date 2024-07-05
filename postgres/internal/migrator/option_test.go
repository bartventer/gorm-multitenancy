package migrator

import (
	"testing"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"
)

func TestOptionFromDB(t *testing.T) {
	var newDB = func() *gorm.DB {
		db, err := gorm.Open(tests.DummyDialector{})
		require.NoError(t, err)
		return db
	}

	tests := []struct {
		name    string
		setup   func(db *gorm.DB) *gorm.DB
		want    option
		wantErr error
	}{
		{
			name: "not found",
			setup: func(db *gorm.DB) *gorm.DB {
				return db // No setup needed, option not set
			},
			wantErr: driver.ErrInvalidMigration,
		},
		{
			name: "not of type option",
			setup: func(db *gorm.DB) *gorm.DB {
				return db.Set(string(migratorKey), "invalid")
			},
			wantErr: driver.ErrInvalidMigration,
		},
		{
			name: "invalid option value",
			setup: func(db *gorm.DB) *gorm.DB {
				return WithOption(option(999))(db)
			},
			wantErr: driver.ErrInvalidMigration,
		},
		{
			name: "valid default option",
			setup: func(db *gorm.DB) *gorm.DB {
				return WithOption(DefaultOption)(db)
			},
			want: DefaultOption,
		},
		{
			name: "valid migrator option",
			setup: func(db *gorm.DB) *gorm.DB {
				return WithOption(MigratorOption)(db)
			},
			want: MigratorOption,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup(newDB())
			got, err := OptionFromDB(db)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
