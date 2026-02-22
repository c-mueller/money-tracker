//go:build integration

package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"icekalt.dev/money-tracker/internal/devmode"
)

func TestPlaygroundAvailable(t *testing.T) {
	if !devmode.Enabled {
		t.Skip("playground only available in dev mode")
	}

	env := setupTestEnv(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/playground", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "GraphQL") {
		t.Error("expected body to contain 'GraphQL'")
	}
}

func TestSwaggerAvailableInDev(t *testing.T) {
	if !devmode.Enabled {
		t.Skip("dev-mode availability test")
	}

	env := setupTestEnv(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/swagger", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "swagger-ui") {
		t.Error("expected body to contain 'swagger-ui'")
	}
}
