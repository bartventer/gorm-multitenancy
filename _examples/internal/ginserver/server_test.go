package ginserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/servertest"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/gin-gonic/gin"
)

// MakeHandler implements [servertest.Harness].
func (c *controller) MakeHandler(ctx context.Context, db *multitenancy.DB) (http.Handler, error) {
	c.db = db

	r := gin.Default()
	c.init(r)

	return r, nil
}

func TestGinServer(t *testing.T) {
	servertest.RunConformance(t, &controller{})
}
