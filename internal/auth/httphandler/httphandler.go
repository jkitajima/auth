package httphandler

import (
	"log/slog"
	"net/http"

	"auth/internal/auth"
	"auth/internal/user"
	userrepo "auth/internal/user/repo/gorm"

	"github.com/go-playground/validator/v10"
	"github.com/jkitajima/composer"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

const (
	Path = "auth/internal/auth/httphandler"
)

type AuthServer struct {
	entity         string
	mux            *chi.Mux
	prefix         string
	service        *auth.Service
	auth           *jwtauth.JWTAuth
	jwtConfig      *auth.JWTConfig
	db             user.Repoer
	inputValidator *validator.Validate
	logger         *slog.Logger
	tracer         trace.Tracer
}

func (s *AuthServer) Prefix() string {
	return s.prefix
}

func (s *AuthServer) Mux() http.Handler {
	return s.mux
}

func (s *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func NewServer(
	jwtauth *jwtauth.JWTAuth,
	jwtconfig *auth.JWTConfig,
	db *gorm.DB,
	validtr *validator.Validate,
	logger *slog.Logger,
	tracer trace.Tracer,
) composer.Server {
	s := &AuthServer{
		entity:         "users",
		prefix:         "/auth",
		mux:            chi.NewRouter(),
		auth:           jwtauth,
		jwtConfig:      jwtconfig,
		db:             userrepo.NewRepo(db, logger),
		inputValidator: validtr,
		logger:         logger,
		tracer:         tracer,
	}

	s.service = &auth.Service{JWTConfig: jwtconfig, UserRepo: s.db}
	s.addRoutes()
	return s
}
