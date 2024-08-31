package irismiddleware_test

import (
	"fmt"

	irismiddleware "github.com/bartventer/gorm-multitenancy/middleware/iris/v8"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/httptest"
)

func ExampleWithTenant() {
	app := iris.New()

	app.Use(irismiddleware.WithTenant(irismiddleware.DefaultWithTenantConfig))

	app.Get("/", func(ctx iris.Context) {
		tenant := ctx.Values().GetString(irismiddleware.TenantKey.String())
		fmt.Println("X-Tenant:", tenant)
		ctx.WriteString("Hello, " + tenant)
	})

	e := httptest.New(nil, app)

	e.GET("/").
		WithHeader(irismiddleware.XTenantHeader, "tenant1").
		Expect().
		Status(httptest.StatusOK).
		Body().IsEqual("Hello, tenant1")

	// Output: X-Tenant: tenant1
}
