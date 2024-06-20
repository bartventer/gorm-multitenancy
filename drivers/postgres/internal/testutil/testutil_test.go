package testutil

import (
	"testing"
)

func TestDeepEqual(t *testing.T) {
	type args struct {
		expected interface{}
		actual   interface{}
		message  []interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name: "Equal integers",
			args: args{
				expected: 10,
				actual:   10,
			},
			want:  true,
			want1: "",
		},
		{
			name: "Unequal integers",
			args: args{
				expected: 10,
				actual:   20,
			},
			want:  false,
			want1: "Expected 10, got 20: Expected 10, got 20",
		},
		{
			name: "Equal strings",
			args: args{
				expected: "Hello",
				actual:   "Hello",
			},
			want:  true,
			want1: "",
		},
		{
			name: "Unequal strings",
			args: args{
				expected: "Hello",
				actual:   "World",
			},
			want:  false,
			want1: "Expected Hello, got World: Expected Hello, got World",
		},
		{
			name: "Custom error message",
			args: args{
				expected: true,
				actual:   false,
				message:  []interface{}{"Custom error message"},
			},
			want:  false,
			want1: "Custom error message: Expected true, got false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := DeepEqual(tt.args.expected, tt.args.actual, tt.args.message...)
			if got != tt.want {
				t.Errorf("DeepEqual() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DeepEqual() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestWithDBName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name     string
		args     args
		existing string
		want     string
	}{
		{
			name: "Replace existing dbname",
			args: args{
				name: "newDB",
			},
			existing: "dbname=oldDB",
			want:     "dbname=newDB",
		},
		{
			name: "Append dbname",
			args: args{
				name: "newDB",
			},
			existing: "host=localhost port=5432 user=postgres password=postgres sslmode=disable",
			want:     "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=newDB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithDBName(tt.args.name)(tt.existing)
			if got != tt.want {
				t.Errorf("WithDBName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTestDB(t *testing.T) {
	type args struct {
		opts []DSNOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				opts: []DSNOption{},
			},
			wantErr: false,
		},
		{
			name: "invalid: ",
			args: args{
				opts: []DSNOption{
					WithDBName("invalid"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("NewTestDB() recovered = %v, wantErr %v", r, tt.wantErr)
					}
				}
			}()
			got := NewTestDB(tt.args.opts...)
			if (got == nil) != tt.wantErr {
				t.Errorf("NewTestDB() error = %v, wantErr %v", got, tt.wantErr)
			}
		})
	}
}
