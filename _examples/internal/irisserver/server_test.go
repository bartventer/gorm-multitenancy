package irisserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/servertest"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/kataras/iris/v12"
)

// MakeHandler implements [servertest.Harness].
func (c *controller) MakeHandler(ctx context.Context, db *multitenancy.DB) (http.Handler, error) {
	c.db = db

	app := iris.New()
	c.init(app)
	app.Build() //https://github.com/kataras/iris/issues/1518#issuecomment-629841233

	return app, nil
}

func TestIrisServer(t *testing.T) {
	servertest.RunConformance(t, &controller{})
}
