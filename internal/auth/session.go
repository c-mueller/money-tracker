package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	SessionName     = "money-tracker-session"
	SessionKeyUser  = "user_id"
	SessionKeyEmail = "email"
	SessionKeyName  = "name"
)

func NewSessionStore(secret string, maxAge int) sessions.Store {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	return store
}
