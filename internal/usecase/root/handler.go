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
		route.Public("GET", "/", h.Get),
	}
}

func (h *RouteHandler) Get(w http.ResponseWriter, _ *http.Request) {
	_ = h.renderer.JSON(w, http.StatusOK, RootResponse{Version: Version})
}
