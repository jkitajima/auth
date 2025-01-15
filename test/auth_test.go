package test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthRegisterUser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	env, err := newEnv()
	if err != nil {
		t.Skip(err)
	}

	route := fmt.Sprintf("http://%s:%s/auth/register", env.host, env.port)
	client := &http.Client{}

	// User email is already used should return 409 Conflict
	t.Run("email_taken", func(t *testing.T) {
		body := strings.NewReader(`
{
    "email": "must_not_touch@email.com",
    "password": "password"
}
		`)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, route, body)
		if err != nil {
			t.Errorf("user: register_user: failed to create request: %v\n", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: register_user: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	// User successful registration should return 201 Created
	t.Run("user_registered", func(t *testing.T) {
		body := strings.NewReader(`
{
	"email": "rogerio.ceni@spfc.com",
	"password": "password"
}
	`)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, route, body)
		if err != nil {
			t.Errorf("user: register_user: failed to create request: %v\n", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: register_user: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestAuthRequestAccessToken(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	env, err := newEnv()
	if err != nil {
		t.Skip(err)
	}

	route := fmt.Sprintf("http://%s:%s/auth/oauth/token", env.host, env.port)
	client := &http.Client{}

	// Unsupported "grant_type" should return 400 Bad Request
	t.Run("unsupported_grant", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("grant_type", "unsupported")
		formData.Set("username", "rogerio.ceni@spfc.com")
		formData.Set("password", "password")
		encodedData := formData.Encode()
		body := strings.NewReader(encodedData)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, route, body)
		if err != nil {
			t.Errorf("user: register_user: failed to create request: %v\n", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: register_user: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Invalid credentials should return 400 Bad Request
	t.Run("invalid_credentials", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("grant_type", "password")
		formData.Set("username", "rogerio.ceni@spfc.com")
		formData.Set("password", "wrong_password")
		encodedData := formData.Encode()
		body := strings.NewReader(encodedData)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, route, body)
		if err != nil {
			t.Errorf("user: register_user: failed to create request: %v\n", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: register_user: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Successful token exchange should return 200 OK
	t.Run("token_exchange", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("grant_type", "password")
		formData.Set("username", "must_not_touch@email.com")
		formData.Set("password", "password")
		encodedData := formData.Encode()
		body := strings.NewReader(encodedData)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, route, body)
		if err != nil {
			t.Errorf("user: register_user: failed to create request: %v\n", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: register_user: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
