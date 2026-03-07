package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"icekalt.dev/money-tracker/ent"
	"icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/config"
	"icekalt.dev/money-tracker/internal/repository"

	_ "modernc.org/sqlite"
)

func setupClient(t *testing.T) *ent.Client {
	t.Helper()
	dbCfg := config.DatabaseConfig{
		Driver: "sqlite",
		DSN:    "file::memory:?cache=shared&_pragma=foreign_keys(1)",
	}
	client, err := repository.NewClient(dbCfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestDBSessionStore_NewSession(t *testing.T) {
	client := setupClient(t)
	store := auth.NewDBSessionStore(client, "test-secret-32-bytes-long-xxxxx", 3600, false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	session, err := store.New(req, auth.SessionName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !session.IsNew {
		t.Error("expected session to be new")
	}
}

func TestDBSessionStore_SaveAndRetrieve(t *testing.T) {
	client := setupClient(t)
	store := auth.NewDBSessionStore(client, "test-secret-32-bytes-long-xxxxx", 3600, false)

	// Create and save a session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, _ := store.New(req, auth.SessionName)
	session.Values[auth.SessionKeyUser] = 42
	session.Values[auth.SessionKeyEmail] = "test@example.com"
	if err := store.Save(req, rec, session); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	// Extract cookie and make a new request
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie set")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}

	session2, err := store.New(req2, auth.SessionName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session2.IsNew {
		t.Error("expected session to not be new")
	}
	if session2.Values[auth.SessionKeyUser] != 42 {
		t.Errorf("expected user_id 42, got %v", session2.Values[auth.SessionKeyUser])
	}
	if session2.Values[auth.SessionKeyEmail] != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", session2.Values[auth.SessionKeyEmail])
	}
}

func TestDBSessionStore_DeleteOnLogout(t *testing.T) {
	client := setupClient(t)
	store := auth.NewDBSessionStore(client, "test-secret-32-bytes-long-xxxxx", 3600, false)

	// Create and save a session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	session, _ := store.New(req, auth.SessionName)
	session.Values[auth.SessionKeyUser] = 42
	if err := store.Save(req, rec, session); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	cookies := rec.Result().Cookies()

	// Now "logout" — set MaxAge to -1 and save
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()

	session2, _ := store.New(req2, auth.SessionName)
	if session2.IsNew {
		t.Fatal("session should exist before logout")
	}

	session2.Options.MaxAge = -1
	if err := store.Save(req2, rec2, session2); err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	// Verify session is gone — replaying the cookie returns a new session
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req3.AddCookie(c)
	}
	session3, _ := store.New(req3, auth.SessionName)
	if !session3.IsNew {
		t.Error("expected session to be new after logout")
	}
}

func TestDBSessionStore_VerifyRowCount(t *testing.T) {
	client := setupClient(t)
	store := auth.NewDBSessionStore(client, "test-secret-32-bytes-long-xxxxx", 3600, false)

	// Create two sessions
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		session, _ := store.New(req, auth.SessionName)
		session.Values[auth.SessionKeyUser] = i + 1
		if err := store.Save(req, rec, session); err != nil {
			t.Fatalf("failed to save session %d: %v", i, err)
		}
	}

	count, err := client.Session.Query().Count(context.Background())
	if err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 session rows, got %d", count)
	}
}
