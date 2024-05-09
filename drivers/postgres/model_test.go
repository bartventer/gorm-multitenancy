package postgres

import (
	"testing"

	"github.com/bartventer/gorm-multitenancy/v5/internal/testutil"
	"gorm.io/gorm"
)

type mockTenantModel struct {
	gorm.Model
	TenantModel
}

func (mockTenantModel) TableName() string {
	return "mock_tenant_model"
}

func TestTenantModelCheckConstraints(t *testing.T) {
	db := testutil.NewTestDB()
	_ = db.AutoMigrate(&mockTenantModel{}) // create table
	t.Cleanup(func() {                     // clean up
		_ = db.Migrator().DropTable(&mockTenantModel{})
	})

	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  *TenantModel
		wantErr bool
	}{
		{
			name: "valid: check constraints",
			args: args{
				db: db,
			},
			fields: &TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "tenant1",
			},
			wantErr: false,
		},
		{
			name: "invalid: schema name too long",
			args: args{
				db: db,
			},
			fields: &TenantModel{
				DomainURL: "tenant1.example.com",
				SchemaName: func() string {
					s := ""
					for i := 0; i < 64; i++ {
						s += "a"
					}
					return s
				}(),
			},
			wantErr: true,
		},
		{
			name: "invalid: schema name too short",
			args: args{
				db: db,
			},
			fields: &TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "a",
			},
			wantErr: true,
		},
		{
			name: "invalid: schema name starts with number",
			args: args{
				db: db,
			},
			fields: &TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "1tenant",
			},
			wantErr: true,
		},
		{
			name: "invalid: schema name starts with pg_",
			args: args{
				db: db,
			},
			fields: &TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "pg_tenant",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockTenantModel{}
			m.DomainURL = tt.fields.DomainURL
			m.SchemaName = tt.fields.SchemaName
			err := tt.args.db.Create(m).Error
			switch {
			case (err != nil) != tt.wantErr:
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			case err != nil:
				// skip delete
				return
			default:
				// delete
				err = tt.args.db.Unscoped().Delete(m).Error
				if err != nil {
					t.Errorf("Delete() error = %v", err)
				}
			}
		})
	}
}
