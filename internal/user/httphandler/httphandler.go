package httphandler

import (
	"log/slog"
	"net/http"

	"auth/internal/user"
	repo "auth/internal/user/repo/gorm"

	"github.com/jkitajima/composer"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

const Path = "auth/internal/user/httphandler"

type UserServer struct {
	entity              string
	mux                 *chi.Mux
	prefix              string
	service             *user.Service
	auth                *jwtauth.JWTAuth
	db                  user.Repoer
	inputValidator      *validator.Validate
	logger              *slog.Logger
	tracer              trace.Tracer
	meter               metric.Meter
	usersDeletedCounter metric.Int64Counter
}

func (s *UserServer) Prefix() string {
	return s.prefix
}

func (s *UserServer) Mux() http.Handler {
	return s.mux
}

func (s *UserServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func NewServer(
	auth *jwtauth.JWTAuth,
	db *gorm.DB,
	validtr *validator.Validate,
	logger *slog.Logger,
	tracer trace.Tracer,
	meter metric.Meter,
) (composer.Server, error) {
	s := &UserServer{
		entity:         "users",
		prefix:         "/users",
		mux:            chi.NewRouter(),
		auth:           auth,
		db:             repo.NewRepo(db, logger),
		inputValidator: validtr,
		logger:         logger,
		tracer:         tracer,
		meter:          meter,
	}
	s.service = &user.Service{Repo: s.db}

	if err := s.instrument(); err != nil {
		return s, err
	}

	s.addRoutes()
	return s, nil
}

func (s *UserServer) instrument() error {
	usersDeletedCounter, err := s.meter.Int64Counter("users_deleted",
		metric.WithDescription("How many new users has been deleted."),
	)
	if err != nil {
		return err
	}
	s.usersDeletedCounter = usersDeletedCounter

	return nil
}
