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

		// otel.Route(r, http.MethodPost, "/", s.handleUserCreate())
		// otel.Route(r, http.MethodPatch, "/{userID}", s.handleUserUpdateByID())
		// otel.Route(r, http.MethodDelete, "/{userID}", s.handleUserSoftDeleteByID())
	})

	// Public routes
	s.mux.Group(func(r chi.Router) {
		otel.Route(r, http.MethodPost, "/register", s.handleUserRegister())
	})
}
