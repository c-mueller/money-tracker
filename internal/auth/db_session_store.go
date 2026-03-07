package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"icekalt.dev/money-tracker/ent"
	entsession "icekalt.dev/money-tracker/ent/session"
)

// DBSessionStore implements sessions.Store with server-side DB storage.
// The cookie contains only a signed session token; all data lives in the DB.
type DBSessionStore struct {
	client *ent.Client
	codecs []securecookie.Codec
	opts   *sessions.Options
}

// NewDBSessionStore creates a DB-backed session store.
func NewDBSessionStore(client *ent.Client, secret string, maxAge int, secure bool) *DBSessionStore {
	codecs := securecookie.CodecsFromPairs([]byte(secret))
	for _, c := range codecs {
		if sc, ok := c.(*securecookie.SecureCookie); ok {
			sc.MaxAge(maxAge)
		}
	}
	return &DBSessionStore{
		client: client,
		codecs: codecs,
		opts: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		},
	}
}

func (s *DBSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *DBSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.opts
	session.Options = &opts
	session.IsNew = true

	cookie, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}

	var token string
	if err := securecookie.DecodeMulti(name, cookie.Value, &token, s.codecs...); err != nil {
		return session, nil
	}

	row, err := s.client.Session.Query().
		Where(entsession.TokenEQ(token)).
		Only(context.Background())
	if err != nil {
		return session, nil
	}

	if row.ExpiresAt.Before(time.Now()) {
		// Expired — delete and return new session
		s.client.Session.DeleteOne(row).Exec(context.Background())
		return session, nil
	}

	if err := gob.NewDecoder(bytes.NewReader(row.Data)).Decode(&session.Values); err != nil {
		return session, nil
	}

	session.ID = token
	session.IsNew = false
	return session, nil
}

func (s *DBSessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		// Delete session from DB
		if session.ID != "" {
			_, err := s.client.Session.Delete().
				Where(entsession.TokenEQ(session.ID)).
				Exec(context.Background())
			if err != nil {
				return err
			}
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	// Encode session values
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(session.Values); err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Duration(session.Options.MaxAge) * time.Second)

	if session.ID == "" {
		// New session — generate token
		token, err := generateToken()
		if err != nil {
			return err
		}
		session.ID = token

		// Extract user_id if present
		var userID *int
		if uid, ok := session.Values[SessionKeyUser].(int); ok {
			userID = &uid
		}

		creator := s.client.Session.Create().
			SetToken(token).
			SetData(buf.Bytes()).
			SetExpiresAt(expiresAt)
		if userID != nil {
			creator = creator.SetUserID(*userID)
		}
		if _, err := creator.Save(context.Background()); err != nil {
			return err
		}
	} else {
		// Existing session — update
		var userID *int
		if uid, ok := session.Values[SessionKeyUser].(int); ok {
			userID = &uid
		}

		updater := s.client.Session.Update().
			Where(entsession.TokenEQ(session.ID)).
			SetData(buf.Bytes()).
			SetExpiresAt(expiresAt)
		if userID != nil {
			updater = updater.SetUserID(*userID)
		} else {
			updater = updater.ClearUserID()
		}
		if _, err := updater.Save(context.Background()); err != nil {
			return err
		}
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
