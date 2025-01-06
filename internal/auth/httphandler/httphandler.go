package httphandler

import (
	"log/slog"
	"net/http"

	"auth/internal/auth"
	"auth/internal/user"
	userRepo "auth/internal/user/repo/gorm"

	"github.com/go-playground/validator/v10"
	"github.com/jkitajima/composer"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

type AuthServer struct {
	mux            *chi.Mux
	prefix         string
	service        *auth.Service
	auth           *jwtauth.JWTAuth
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
	db *gorm.DB,
	validtr *validator.Validate,
	logger *slog.Logger,
	tracer trace.Tracer,
) composer.Server {
	s := &AuthServer{
		prefix:         "/auth",
		mux:            chi.NewRouter(),
		auth:           jwtauth,
		db:             userRepo.NewRepo(db, logger),
		inputValidator: validtr,
		logger:         logger,
		tracer:         tracer,
	}

	s.service = &auth.Service{UserRepo: s.db}
	s.addRoutes()
	return s
}

const (
	Path string = "auth/internal/auth/httphandler"
)
