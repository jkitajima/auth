package httphandler

import (
	"net/http"

	"auth/pkg/otel"

	"github.com/jkitajima/responder"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func (s *AuthServer) addRoutes() {
	// Private routes
	s.mux.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(s.auth))
		r.Use(responder.RespondAuth(s.auth))
	})

	// Public routes
	s.mux.Group(func(r chi.Router) {
		otel.Route(r, http.MethodPost, "/oauth/token", s.handleRequestAccessToken())
		otel.Route(r, http.MethodPost, "/register", s.handleUserRegister())
	})
}
