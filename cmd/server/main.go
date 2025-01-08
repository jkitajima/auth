package main

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
	userrepo "auth/internal/user/repo/gorm"

	servercomposer "github.com/jkitajima/composer"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	service string = "auth"
	path    string = service + "/cmd/server"
)

func main() {
	ctx := context.Background()
	if err := exec(ctx, os.Args, os.Stdin, os.Stdout, os.Stderr, os.Getenv, os.Getwd); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func exec(
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

	db, err := initDB(cfg.DB)
	if err != nil {
		return err
	}

	inputValidator := validator.New(validator.WithRequiredStructEnabled())
	logger := otelslog.NewLogger(service)
	tracer := otel.Tracer(service)

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
	healthCheck := SetupHealthCheck(cfg, logger)
	authServer := authserver.NewServer(jwtAuth, (*auth.JWTConfig)(cfg.Auth.JWT), db, inputValidator, logger, tracer)
	if err := composer.Compose(healthCheck, authServer); err != nil {
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

	// Add HTTP instrumentation for the whole server
	otelhandler := otelhttp.NewHandler(composer.Mux, "/")

	// Server config
	server := &http.Server{
		Addr:           net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		BaseContext:    func(net.Listener) context.Context { return ctx },
		WriteTimeout:   time.Second * time.Duration(cfg.Server.Timeout.Write),
		ReadTimeout:    time.Second * time.Duration(cfg.Server.Timeout.Read),
		IdleTimeout:    time.Second * time.Duration(cfg.Server.Timeout.Idle),
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
		Handler:        otelhandler,
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

func initDB(config *DB) (*gorm.DB, error) {
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

	return db, nil
}
