package driver

import (
	"testing"
)

type mockTenantTabler struct {
	tableName   string
	sharedModel bool
}

func (m *mockTenantTabler) TableName() string   { return m.tableName }
func (m *mockTenantTabler) IsSharedModel() bool { return m.sharedModel }

func TestModelsToInterfaces(t *testing.T) {
	tests := []struct {
		name   string
		models []TenantTabler
		want   int // Expected length of the result slice.
	}{
		{
			name:   "empty slice",
			models: []TenantTabler{},
			want:   0,
		},
		{
			name: "single model",
			models: []TenantTabler{
				&mockTenantTabler{tableName: "test_table", sharedModel: false},
			},
			want: 1,
		},
		{
			name: "multiple models",
			models: []TenantTabler{
				&mockTenantTabler{tableName: "test_table1", sharedModel: false},
				&mockTenantTabler{tableName: "test_table2", sharedModel: true},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelsToInterfaces(tt.models)
			if len(got) != tt.want {
				t.Errorf("ModelsToInterfaces() got = %v, want %v", len(got), tt.want)
			}
		})
	}
}
