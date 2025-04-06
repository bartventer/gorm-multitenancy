package nethttpserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/bartventer/gorm-multitenancy/examples/v8/internal/servertest"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/urfave/negroni"
)

// MakeHandler implements [servertest.Harness].
func (c *controller) MakeHandler(ctx context.Context, db *multitenancy.DB) (http.Handler, error) {
	c.db = db

	mux := http.NewServeMux()
	n := negroni.Classic()
	c.init(mux, n)

	return n, nil
}

func TestNetHTTPServer(t *testing.T) {
	servertest.RunConformance(t, &controller{})
}
