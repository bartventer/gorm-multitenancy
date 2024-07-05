package namespace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateTenantName(t *testing.T) {
	tests := []struct {
		tenantName string
		wantErr    bool
	}{
		{"valid_name", false},
		{"_validName123", false},
		{"in", true},
		{"pg_invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.tenantName, func(t *testing.T) {
			err := Validate(tt.tenantName)
			if tt.wantErr {
				require.Error(t, err, "Validate() should fail")
			} else {
				require.NoError(t, err, "Validate() should not fail")
			}
		})
	}
}
