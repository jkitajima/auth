package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth/internal/auth"
	authserver "auth/internal/auth/httphandler"
	userserver "auth/internal/user/httphandler"
	userrepo "auth/internal/user/repo/gorm"

	servercomposer "github.com/jkitajima/composer"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	Service string = "auth"
	Path    string = Service + "/internal/server"
)

func Exec(
	ctx context.Context,
	args []string,
	_ io.Reader,
	stdout io.Writer,
	_ io.Writer,
	_ func(string) string,
	_ func() (string, error),
) error {
	cfg, err := NewConfig(stdout, args)
	if err != nil {
		return err
	}

	// Setting up dependencies
	jwtAuth := jwtauth.New(cfg.Auth.JWT.Algorithm, []byte(cfg.Auth.JWT.Key), nil)

	db, err := initDB(cfg.Environment, cfg.DB)
	if err != nil {
		return err
	}

	inputValidator := validator.New(validator.WithRequiredStructEnabled())
	logger := otelslog.NewLogger(Service)
	tracer := otel.Tracer(Service)
	meter := otel.Meter(Service)

	// Mounting routers
	composer := servercomposer.NewComposer(
		middleware.Recoverer,
		middleware.AllowContentType(
			"application/json",
			"application/x-www-form-urlencoded",
		),
		middleware.CleanPath,
		middleware.RedirectSlashes,
	)

	// API Docs
	composer.Mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		// Path relative to the server binary
		http.ServeFile(w, r, "./api/swagger.html") // if binary is at root
	})

	healthCheck := SetupHealthCheck(cfg, logger)

	authServer, err := authserver.NewServer(jwtAuth, (*auth.JWTConfig)(cfg.Auth.JWT), db, inputValidator, logger, tracer, meter)
	if err != nil {
		return err
	}

	userServer, err := userserver.NewServer(jwtAuth, db, inputValidator, logger, tracer, meter)
	if err != nil {
		return err
	}

	if err := composer.Compose(healthCheck, authServer, userServer); err != nil {
		return err
	}

	// Set up Instrumentation
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	// Server config
	server := &http.Server{
		Addr:           net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		BaseContext:    func(net.Listener) context.Context { return ctx },
		WriteTimeout:   time.Second * time.Duration(cfg.Server.Timeout.Write),
		ReadTimeout:    time.Second * time.Duration(cfg.Server.Timeout.Read),
		IdleTimeout:    time.Second * time.Duration(cfg.Server.Timeout.Idle),
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
		Handler:        composer,
	}

	// Graceful shutdown
	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverChan := make(chan error, 1)
	go func() {
		<-notifyCtx.Done()

		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Server.Timeout.Shutdown))
		defer cancel()

		if err := server.Shutdown(timeoutCtx); err != nil {
			serverChan <- err
		}
		serverChan <- nil
	}()

	log.Printf("server listening on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return <-serverChan
}

func initDB(env Environment, config *DB) (*gorm.DB, error) {
	config.DSN = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host,
		config.User,
		config.Password,
		config.Name,
		config.Port,
		config.SSL,
	)
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return &gorm.DB{}, err
	}

	// UUID support for PostgreSQL
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)

	// Migrate the schema
	db.AutoMigrate(&userrepo.UserModel{})

	// Seeding data for tests
	if env == EnvironmentTest {
		// Test users passwords: "password"
		db.Exec(`INSERT INTO "User" (id, email, password, created_at, updated_at) VALUES ('794defc3-109a-4c6f-a7d2-cb976065ea80', 'to_be_deleted@email.com', '$argon2id$v=19$m=65536,t=1,p=8$c8RKqOjWl12MOm0kUIzY1g$zPDSs37yzKyh6SwQFpkmdAq+hf1PzglTIAIGGcsj8ro', '2020-07-04 11:05:21.775', '2020-07-04 11:05:21.775');`)
		db.Exec(`INSERT INTO "User" (id, email, password, created_at, updated_at) VALUES ('1aef49bd-3296-45fb-84b9-083cf81b0e44', 'must_not_touch@email.com', '$argon2id$v=19$m=65536,t=1,p=8$c8RKqOjWl12MOm0kUIzY1g$zPDSs37yzKyh6SwQFpkmdAq+hf1PzglTIAIGGcsj8ro', '2020-07-04 11:05:21.775', '2020-07-04 11:05:21.775');`)
	}

	return db, nil
}
