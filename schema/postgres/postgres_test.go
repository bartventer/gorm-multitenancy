package postgres

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Test_getSchemaNameFromSqlExpr(t *testing.T) {
	type args struct {
		tableExprSql string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with schema and table",
			args: args{
				tableExprSql: "\"test_domain\".\"mock_private\"",
			},
			want: "test_domain",
		},
		{
			name: "Test with only table",
			args: args{
				tableExprSql: "\"mock_private\"",
			},
			want: "",
		},
		{
			name: "Test with empty string",
			args: args{
				tableExprSql: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSchemaNameFromSqlExpr(tt.args.tableExprSql); got != tt.want {
				t.Errorf("getSchemaNameFromTableExpreSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestGetSchemaNameFromDb(t *testing.T) {
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test with schema",
			args: args{
				tx: &gorm.DB{Statement: &gorm.Statement{TableExpr: &clause.Expr{SQL: "\"schema\".table"}}},
			},
			want:    "schema",
			wantErr: false,
		},
		{
			name: "Test without schema",
			args: args{
				tx: &gorm.DB{Statement: &gorm.Statement{TableExpr: &clause.Expr{SQL: "table"}}},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Test with nil TableExpr",
			args: args{
				tx: &gorm.DB{Statement: &gorm.Statement{}},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSchemaNameFromDb(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSchemaNameFromDb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSchemaNameFromDb() = %v, want %v", got, tt.want)
			}
		})
	}
}
