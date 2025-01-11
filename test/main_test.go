package test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"auth/internal/server"

	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup dependencies
	compose, err := tc.NewDockerCompose("../compose.yaml")
	if err != nil {
		log.Fatalf("testcontainers: docker compose init failed: %v\n", err)
		return
	}

	defer func() {
		if err := compose.Down(ctx, tc.RemoveOrphans(true), tc.RemoveImagesLocal); err != nil {
			log.Fatalf("testcontainers: docker compose down failed: %v\n", err)
			return
		}
	}()

	if err := compose.Up(ctx, tc.Wait(true)); err != nil {
		log.Fatalf("testcontainers: docker compose up failed: %v\n", err)
		return
	}

	// Run server
	args := []string{
		"auth",
		"--env", "test",
		"--config", "env.test.yaml",
	}

	// Required env vars for testing
	os.Setenv("AUTH_SERVER_HOST", "localhost")
	os.Setenv("AUTH_SERVER_PORT", "8111")

	go server.Exec(ctx, args, nil, os.Stdout, nil, nil, nil)

	// Wait for server readiness
	if err := waitForReadiness(ctx); err != nil {
		log.Fatal(err)
		return
	}

	m.Run()
}

type env struct {
	host string
	port string
}

func newEnv() (*env, error) {
	host := os.Getenv("AUTH_SERVER_HOST")
	if host == "" {
		return &env{}, errors.New(`env var "AUTH_SERVER_HOST" is not set`)
	}

	port := os.Getenv("AUTH_SERVER_PORT")
	if port == "" {
		return &env{}, errors.New(`env var "AUTH_SERVER_PORT" is not set`)
	}

	return &env{
		host: host,
		port: port,
	}, nil
}

func waitForReadiness(ctx context.Context) error {
	env, err := newEnv()
	if err != nil {
		return err
	}

	client := &http.Client{}
	url := fmt.Sprintf("http://%s:%s/healthz/readiness", env.host, env.port)
	startTime := time.Now()
	waitTime := 750 * time.Millisecond
	timeout := 5 * time.Second // minimum of server.health.delay + 1

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %s\n", err.Error())
			goto TimeoutCheck
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

	TimeoutCheck:
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for server readiness")
			}
			// wait a little while between checks
			time.Sleep(waitTime)
		}
	}
}
