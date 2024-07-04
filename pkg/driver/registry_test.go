package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock models for testing.
type sharedModel struct{}

func (sharedModel) TableName() string   { return "public.shared" }
func (sharedModel) IsSharedModel() bool { return true }

type tenantModel struct{}

func (tenantModel) TableName() string   { return "tenant_specific" }
func (tenantModel) IsSharedModel() bool { return false }

type invalidSharedModel struct{}

func (invalidSharedModel) TableName() string   { return "shared" }
func (invalidSharedModel) IsSharedModel() bool { return true }

type invalidTenantModel struct{}

func (invalidTenantModel) TableName() string   { return "public.tenant_specific" }
func (invalidTenantModel) IsSharedModel() bool { return false }

func TestValidateSharedModel(t *testing.T) {
	require.NoError(t, validateSharedModel("public.shared"), "valid shared model should not return an error")
	assert.Error(t, validateSharedModel("shared"), "invalid shared model should return an error")
}

func TestValidateTenantModel(t *testing.T) {
	require.NoError(t, validateTenantModel("tenant_specific"), "valid tenant model should not return an error")
	assert.Error(t, validateTenantModel("public.tenant_specific"), "invalid tenant model should return an error")
}

func TestNewConfig(t *testing.T) {
	t.Run("valid models", func(t *testing.T) {
		config, err := NewModelRegistry([]TenantTabler{sharedModel{}, tenantModel{}}...)
		require.NoError(t, err)
		assert.Len(t, config.SharedModels, 1, "there should be one shared model")
		assert.Len(t, config.TenantModels, 1, "there should be one tenant model")
	})

	t.Run("invalid table names", func(t *testing.T) {
		_, err := NewModelRegistry([]TenantTabler{invalidSharedModel{}, invalidTenantModel{}}...)
		require.Error(t, err, "invalid table names should return an error")
	})
}
