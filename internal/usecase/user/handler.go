//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=./openapi/config.yaml ./openapi/spec.yaml

package user

import (
	"encoding/json"
	"net/http"

	"github.com/gookit/validate"
	"github.com/hpcsc/chi-api/internal/response"
	"github.com/hpcsc/chi-api/internal/route"
	"github.com/unrolled/render"
)

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
		route.Protected("POST", "/users", h.PostApiUsers),
	}
}

func (h *RouteHandler) PostApiUsers(w http.ResponseWriter, req *http.Request) {
	var postData CreateUserRequest
	if err := json.NewDecoder(req.Body).Decode(&postData); err != nil {
		_ = h.renderer.JSON(w, http.StatusBadRequest, response.Fail("received invalid request body"))
		return
	}

	v := validate.Struct(postData)
	if !v.Validate() {
		_ = h.renderer.JSON(w, http.StatusBadRequest, response.FailWithValidationErrors(v.Errors))
		return
	}

	// do something with postData

	_ = h.renderer.JSON(w, http.StatusOK, response.Succeed())
}
