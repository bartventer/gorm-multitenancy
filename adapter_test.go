package multitenancy

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type adapter struct{}

var _ Adapter = new(adapter)

func (m *adapter) AdaptDB(ctx context.Context, db *gorm.DB) (*DB, error) {
	if db.Name() == "err" {
		return nil, errors.New("forced error")
	}
	return nil, nil
}

// OpenDBURL implements Adapter.
func (m *adapter) OpenDBURL(ctx context.Context, u *url.URL, opts ...gorm.Option) (*DB, error) {
	if u.Scheme == "err" {
		return nil, errors.New("forced error")
	}
	return nil, nil
}

func TestAdaptDB(t *testing.T) {
	ctx := context.Background()
	mux := new(driverMux)

	fake := &adapter{}
	mux.Register("foo", fake)
	mux.Register("err", fake)

	for _, tc := range []struct {
		name    string
		driver  string
		wantErr bool
	}{
		{
			name:    "unregistered driver",
			driver:  "bar",
			wantErr: true,
		},
		{
			name:    "driver returns error",
			driver:  "err",
			wantErr: true,
		},
		{
			name:    "valid driver",
			driver:  "foo",
			wantErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &gorm.DB{} // Mock gorm.DB with the necessary driver name setup.
			db.Config = &gorm.Config{Dialector: gorm.Dialector(&mockDialector{name: tc.driver})}
			_, gotErr := mux.AdaptDB(ctx, db)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v, want error %v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestOpenDBURL(t *testing.T) {
	ctx := context.Background()
	mux := new(driverMux)

	fake := &adapter{}
	mux.Register("foo", fake)
	mux.Register("err", fake)

	for _, tc := range []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     ":foo",
			wantErr: true,
		},
		{
			name:    "invalid URL no scheme",
			url:     "foo",
			wantErr: true,
		},
		{
			name:    "unregistered scheme",
			url:     "bar://mydb",
			wantErr: true,
		},
		{
			name:    "func returns error",
			url:     "err://mydb",
			wantErr: true,
		},
		{
			name: "no query options",
			url:  "foo://mydb",
		},
		{
			name: "empty query options",
			url:  "foo://mydb?",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, gotErr := mux.OpenDB(ctx, tc.url)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v, want error %v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	fake := &adapter{}

	// Test registering a new driver.
	Register("new", fake)

	// Test registering an existing driver, should panic.
	assert.Panicsf(t, func() { Register("new", fake) }, "Register() did not panic")
}

func TestOpen(t *testing.T) {
	fake := &adapter{}
	Register("foo", fake)

	// Test creating a new DB instance with a registered driver.
	_, err := Open(gorm.Dialector(&mockDialector{name: "foo"}))
	require.NoError(t, err)

	// Test creating a new DB instance with an unregistered driver, should return an error.
	_, err = Open(gorm.Dialector(&mockDialector{name: "bar"}))
	assert.Error(t, err)
}

func TestOpenDB(t *testing.T) {
	fake := &adapter{}
	Register("foo2", fake)

	// Test creating a new DB instance with a registered driver.
	_, err := OpenDB(context.Background(), "foo2://mydb")
	require.NoError(t, err)

	// Test creating a new DB instance with an unregistered driver, should return an error.
	_, err = OpenDB(context.Background(), "bar://mydb")
	assert.Error(t, err)
}

// mockDialector is a mock implementation of gorm.Dialector for testing purposes.
var _ gorm.Dialector = new(mockDialector)

type mockDialector struct {
	name string
}

// BindVarTo implements [gorm.Dialector].
func (m *mockDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}

// DataTypeOf implements [gorm.Dialector].
func (m *mockDialector) DataTypeOf(*schema.Field) string {
	return ""
}

// DefaultValueOf implements [gorm.Dialector].
func (m *mockDialector) DefaultValueOf(*schema.Field) clause.Expression {
	return nil
}

// Explain implements [gorm.Dialector].
func (m *mockDialector) Explain(sql string, vars ...interface{}) string {
	return ""
}

// Initialize implements [gorm.Dialector].
func (m *mockDialector) Initialize(*gorm.DB) error {
	return nil
}

// Migrator implements [gorm.Dialector].
func (m *mockDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return nil
}

// Name implements [gorm.Dialector].
func (m *mockDialector) Name() string {
	return m.name
}

// QuoteTo implements [gorm.Dialector].
func (m *mockDialector) QuoteTo(clause.Writer, string) {}
