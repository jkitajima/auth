package httphandler

import (
	"net/http"

	"auth/pkg/otel"

	"github.com/jkitajima/responder"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func (s *UserServer) addRoutes() {
	// Private routes
	s.mux.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(s.auth))
		r.Use(responder.RespondAuth(s.auth))

		otel.Route(r, http.MethodPost, "/{userID}/delete", s.handleUserHardDeleteByID())
	})

	// Public routes
	// s.mux.Group(func(r chi.Router) {
	// })
}
