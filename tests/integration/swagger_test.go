//go:build integration

package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"icekalt.dev/money-tracker/internal/auth"
)

func TestOpenAPISpecNoAuth(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/openapi.yaml", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "yaml") {
		t.Errorf("expected Content-Type to contain 'yaml', got %q", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "openapi:") {
		t.Error("expected body to contain 'openapi:'")
	}
}

func TestSwaggerUINoAuth(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/swagger", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestSwaggerUIWithSession(t *testing.T) {
	env := setupTestEnv(t)

	// Create a valid session cookie
	fakeReq := httptest.NewRequest(http.MethodGet, "/", nil)
	fakeRec := httptest.NewRecorder()
	session, _ := env.sessionStore.New(fakeReq, auth.SessionName)
	session.Values[auth.SessionKeyUser] = env.userID
	session.Save(fakeReq, fakeRec)

	cookies := fakeRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie was set")
	}

	req, _ := http.NewRequest("GET", env.server.URL+"/swagger", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

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

func TestSwaggerUIWithToken(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("GET", env.server.URL+"/swagger", nil)
	req.Header.Set("Authorization", "Bearer "+env.token)

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
