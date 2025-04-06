package echoserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/servertest"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/labstack/echo/v4"
)

// MakeHandler implements [servertest.Harness].
func (c *controller) MakeHandler(ctx context.Context, db *multitenancy.DB) (http.Handler, error) {
	c.db = db

	e := echo.New()
	c.init(e)

	return e, nil
}

func TestEchoServer(t *testing.T) {
	servertest.RunConformance(t, &controller{})
}
