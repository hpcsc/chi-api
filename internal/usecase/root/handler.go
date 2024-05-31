//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=./openapi/config.yaml ./openapi/spec.yaml

package root

import (
	"net/http"

	"github.com/hpcsc/chi-api/internal/route"
	"github.com/unrolled/render"
)

var Version = "main"

var _ ServerInterface = (*RouteHandler)(nil)
var _ route.Routable = (*RouteHandler)(nil)

func NewHandler() route.Routable {
	return &RouteHandler{
		renderer: render.New(),
	}
}

type RouteHandler struct {
	renderer *render.Render
}

func (h *RouteHandler) Routes() []*route.Route {
	return []*route.Route{
		route.Public("GET", "/", h.GetApi),
	}
}

func (h *RouteHandler) GetApi(w http.ResponseWriter, _ *http.Request) {
	_ = h.renderer.JSON(w, http.StatusOK, RootResponse{Version: Version})
}
