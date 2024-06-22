package schema

import (
	"testing"
)

func Test_getSchemaNameFromSqlExpr(t *testing.T) {
	type args struct {
		tableExprSQL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with schema and table",
			args: args{
				tableExprSQL: "\"test_domain\".\"mock_private\"",
			},
			want: "test_domain",
		},
		{
			name: "Test with only table",
			args: args{
				tableExprSQL: "\"mock_private\"",
			},
			want: "",
		},
		{
			name: "Test with empty string",
			args: args{
				tableExprSQL: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSchemaNameFromSQLExpr(tt.args.tableExprSQL); got != tt.want {
				t.Errorf("getSchemaNameFromTableExpreSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
