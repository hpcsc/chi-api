package usecase

import (
	"github.com/go-chi/chi/v5"
	"github.com/hpcsc/chi-api/internal/usecase/root"
	"github.com/hpcsc/chi-api/internal/usecase/user"
)

func Register(router chi.Router) {
	root.Register(router)
	user.Register(router)
}
