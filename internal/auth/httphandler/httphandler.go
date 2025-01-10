package httphandler

import (
	"log/slog"
	"net/http"

	"auth/internal/auth"
	"auth/internal/user"
	userrepo "auth/internal/user/repo/gorm"

	"github.com/go-playground/validator/v10"
	"github.com/jkitajima/composer"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

const (
	Path = "auth/internal/auth/httphandler"
)

type AuthServer struct {
	entity                 string
	mux                    *chi.Mux
	prefix                 string
	service                *auth.Service
	auth                   *jwtauth.JWTAuth
	jwtConfig              *auth.JWTConfig
	db                     user.Repoer
	inputValidator         *validator.Validate
	logger                 *slog.Logger
	tracer                 trace.Tracer
	meter                  metric.Meter
	usersCreatedCounter    metric.Int64Counter
	tokensGeneratedCounter metric.Int64Counter
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
	meter metric.Meter,
) (composer.Server, error) {
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
		meter:          meter,
	}
	s.service = &auth.Service{JWTConfig: jwtconfig, UserRepo: s.db}

	if err := s.instrument(); err != nil {
		return s, err
	}

	s.addRoutes()
	return s, nil
}

func (s *AuthServer) instrument() error {
	usersCreatedCounter, err := s.meter.Int64Counter("users_registered",
		metric.WithDescription("How many new users has been successfully registered."),
	)
	if err != nil {
		return err
	}
	s.usersCreatedCounter = usersCreatedCounter

	tokensGeneratedCounter, err := s.meter.Int64Counter("tokens_generated",
		metric.WithDescription("How many tokens was generated after exchange flow."),
	)
	if err != nil {
		return err
	}
	s.tokensGeneratedCounter = tokensGeneratedCounter

	return nil
}
