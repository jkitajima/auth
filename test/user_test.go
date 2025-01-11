package test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserHardDeleteByID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	env, err := newEnv()
	if err != nil {
		t.Skip(err)
	}

	route := fmt.Sprintf("http://%s:%s/users/{id}/delete", env.host, env.port)
	client := &http.Client{}
	const token = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiaHR0cDovL2xvY2FsaG9zdDo4MTExLyJdLCJleHAiOjQ3NjYxMzk0NjcsImlhdCI6MTczNjYwOTg2NywiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo4MTExLyIsImp0aSI6ImZmMTBkNTkzLThmODgtNGU3Yy1hMDMwLTg1MjI0MWU2MmZlYyIsIm5iZiI6MTczNjYwOTg2Nywic3ViIjoiNzk0ZGVmYzMtMTA5YS00YzZmLWE3ZDItY2I5NzYwNjVlYTgwIn0.vSztx6LAOej9kVj3MflrI7S6aU1YfET1FnhpvyNKSuo"

	// Anonymous requests receives Unauthorized response
	t.Run("unauthorized", func(t *testing.T) {
		url := strings.Replace(route, "{id}", "anonymous", 1)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: failed to create request: %v\n", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Malformatted ID receives Bad Request
	t.Run("malformatted_id", func(t *testing.T) {
		url := strings.Replace(route, "{id}", "malformatted_id", 1)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: failed to create request: %v\n", err)
		}

		req.Header.Set("Authorization", token)

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Invalid password receives Bad Request
	t.Run("invalid_password", func(t *testing.T) {
		url := strings.Replace(route, "{id}", "794defc3-109a-4c6f-a7d2-cb976065ea80", 1)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: failed to create request: %v\n", err)
		}

		req.Header.Set("Authorization", token)

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// User tries to delete another user receives Forbidden
	t.Run("forbidden", func(t *testing.T) {
		url := strings.Replace(route, "{id}", "1aef49bd-3296-45fb-84b9-083cf81b0e44", 1)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: failed to create request: %v\n", err)
		}

		req.Header.Set("Authorization", token)

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Since this test actually deletes the user, it should be placed at last
	// User tries to delete another user receives Forbidden
	t.Run("deleted", func(t *testing.T) {
		body := strings.NewReader(`
{
	"password": "password"
}
		`)

		url := strings.Replace(route, "{id}", "794defc3-109a-4c6f-a7d2-cb976065ea80", 1)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: failed to create request: %v\n", err)
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("user: hard_delete_by_id: invalid id format: request failed: %v\n", err)
		}
		defer resp.Body.Close()

		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
