package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/config"
	mw "icekalt.dev/money-tracker/internal/middleware"
	"icekalt.dev/money-tracker/internal/repository"
	"icekalt.dev/money-tracker/internal/service"

	_ "modernc.org/sqlite"
)

type testSetup struct {
	tokenSvc *service.APITokenService
	token    string
	userID   int
}

func setup(t *testing.T) *testSetup {
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

	userRepo := repository.NewUserRepository(client)
	tokenRepo := repository.NewAPITokenRepository(client)

	userSvc := service.NewUserService(userRepo)
	tokenSvc := service.NewAPITokenService(tokenRepo)

	user, err := userSvc.GetOrCreate(context.Background(), "test-sub", "test@example.com", "Test")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	userCtx := service.WithUserID(context.Background(), user.ID)
	plainToken, _, err := tokenSvc.Create(userCtx, "test-token")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	return &testSetup{
		tokenSvc: tokenSvc,
		token:    plainToken,
		userID:   user.ID,
	}
}

func TestAuthMiddleware_BearerToken(t *testing.T) {
	ts := setup(t)
	sessionStore := auth.NewSessionStore("test-secret-for-middleware-tests", 3600)
	authMiddleware := mw.Auth(sessionStore, ts.tokenSvc, 0)

	e := echo.New()
	handler := authMiddleware(func(c echo.Context) error {
		userID := c.Get(mw.UserIDContextKey)
		if userID != ts.userID {
			t.Errorf("expected userID %d, got %v", ts.userID, userID)
		}
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+ts.token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	ts := setup(t)
	sessionStore := auth.NewSessionStore("test-secret-for-middleware-tests", 3600)
	authMiddleware := mw.Auth(sessionStore, ts.tokenSvc, 0)

	e := echo.New()
	handler := authMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-value")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestAuthMiddleware_NoAuth(t *testing.T) {
	ts := setup(t)
	sessionStore := auth.NewSessionStore("test-secret-for-middleware-tests", 3600)
	authMiddleware := mw.Auth(sessionStore, ts.tokenSvc, 0)

	e := echo.New()
	handler := authMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for no auth")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestAuthMiddleware_SessionCookie(t *testing.T) {
	ts := setup(t)
	sessionStore := auth.NewSessionStore("test-secret-for-middleware-tests", 3600)
	authMiddleware := mw.Auth(sessionStore, ts.tokenSvc, 0)

	e := echo.New()
	handler := authMiddleware(func(c echo.Context) error {
		userID := c.Get(mw.UserIDContextKey)
		if userID != ts.userID {
			t.Errorf("expected userID %d, got %v", ts.userID, userID)
		}
		return c.String(http.StatusOK, "ok")
	})

	// Create a valid session by writing to the store
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Save a session
	session, _ := sessionStore.New(req, auth.SessionName)
	session.Values[auth.SessionKeyUser] = ts.userID
	session.Save(req, rec)

	// Extract the cookie
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie was set")
	}

	// Make a new request with the session cookie
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}
}

func TestTokenOnlyAuth_BearerToken(t *testing.T) {
	ts := setup(t)
	tokenAuthMW := mw.TokenOnlyAuth(ts.tokenSvc, 0)

	e := echo.New()
	handler := tokenAuthMW(func(c echo.Context) error {
		userID := c.Get(mw.UserIDContextKey)
		if userID != ts.userID {
			t.Errorf("expected userID %d, got %v", ts.userID, userID)
		}
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+ts.token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTokenOnlyAuth_InvalidToken(t *testing.T) {
	ts := setup(t)
	tokenAuthMW := mw.TokenOnlyAuth(ts.tokenSvc, 0)

	e := echo.New()
	handler := tokenAuthMW(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestTokenOnlyAuth_NoToken(t *testing.T) {
	ts := setup(t)
	tokenAuthMW := mw.TokenOnlyAuth(ts.tokenSvc, 0)

	e := echo.New()
	handler := tokenAuthMW(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// No auth header at all
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error when no token provided")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestTokenOnlyAuth_SessionCookieRejected(t *testing.T) {
	ts := setup(t)
	sessionStore := auth.NewSessionStore("test-secret-for-middleware-tests", 3600)
	tokenAuthMW := mw.TokenOnlyAuth(ts.tokenSvc, 0)

	e := echo.New()
	handler := tokenAuthMW(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Create a valid session cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	session, _ := sessionStore.New(req, auth.SessionName)
	session.Values[auth.SessionKeyUser] = ts.userID
	session.Save(req, rec)

	cookies := rec.Result().Cookies()

	// Send request with session cookie but no Bearer token
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	c := e.NewContext(req2, rec2)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error — TokenOnlyAuth should reject session cookies")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}
