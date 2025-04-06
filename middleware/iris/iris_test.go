package irismiddleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/httptest"
)

func TestWithTenant(t *testing.T) {
	type args struct {
		tenant string
		config WithTenantConfig
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid: tenant from host",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "valid: tenant from header",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "test-with-skipper",
			args: args{
				config: WithTenantConfig{
					Skipper: func(ctx iris.Context) bool {
						return true
					},
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "test-with-error-handler",
			args: args{
				config: WithTenantConfig{
					TenantGetters: []func(ctx iris.Context) (string, error){
						func(ctx iris.Context) (string, error) {
							return "", errors.New("forced error")
						},
					},
					ErrorHandler: func(ctx iris.Context, err error) {
						ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "forced error"})
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "test-with-success-handler",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{
					TenantGetters: []func(ctx iris.Context) (string, error){
						func(ctx iris.Context) (string, error) {
							return "tenant1", nil
						},
					},
					SuccessHandler: func(ctx iris.Context) {
						t.Log("success")
					},
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := iris.New()
			app.Use(WithTenant(tt.args.config))
			app.Get("/", func(ctx iris.Context) {
				tenant := ctx.Values().GetString(TenantKey.String())
				if tt.args.config.Skipper != nil && tt.args.config.Skipper(ctx) {
					ctx.WriteString("")
					return
				}

				if tt.args.config.SuccessHandler == nil && tenant != tt.want {
					ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "expected tenant " + tt.want + ", got " + tenant})
					return
				}
				ctx.WriteString(tenant)
			})

			e := httptest.New(t, app)

			req := e.GET("/").
				WithHost(tt.args.tenant+".example.com").
				WithHeader(nethttp.XTenantHeader, tt.args.tenant)

			if tt.wantErr {
				req.Expect().Status(httptest.StatusInternalServerError)
			} else {
				req.Expect().Status(httptest.StatusOK).Body().IsEqual(tt.want)
			}
		})
	}
}
