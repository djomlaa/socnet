package handler

import (
	"github.com/matryer/way"
	"net/http"
	"github.com/djomlaa/socnet/internal/service"
)

type handler struct {
	*service.Service
}

// New creates predefined routing.
func New(s *service.Service) http.Handler {

	h := &handler{s}

	api := way.NewRouter()
	api.HandleFunc("POST", "/login", h.login)
	api.HandleFunc("GET", "/auth_user", h.authUser)
	api.HandleFunc("POST", "/users", h.createUser)
	api.HandleFunc("POST", "/users/:username/toggle_follow", h.toggleFollow)

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(api)))

	return r
}
