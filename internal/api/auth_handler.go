package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/auth"
)

type AuthHandler struct {
	oidcCfg *auth.OIDCConfig
	store   sessions.Store
	services *Services
}

func NewAuthHandler(oidcCfg *auth.OIDCConfig, store sessions.Store, services *Services) *AuthHandler {
	return &AuthHandler{
		oidcCfg:  oidcCfg,
		store:    store,
		services: services,
	}
}

func (h *AuthHandler) HandleLogin(c echo.Context) error {
	if h.oidcCfg == nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "OIDC not configured"})
	}

	state, err := generateState()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate state"})
	}

	session, _ := h.store.Get(c.Request(), auth.SessionName)
	session.Values["oauth_state"] = state
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save session"})
	}

	return c.Redirect(http.StatusFound, h.oidcCfg.OAuth2Config.AuthCodeURL(state))
}

func (h *AuthHandler) HandleCallback(c echo.Context) error {
	if h.oidcCfg == nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "OIDC not configured"})
	}

	session, _ := h.store.Get(c.Request(), auth.SessionName)
	savedState, _ := session.Values["oauth_state"].(string)
	if c.QueryParam("state") != savedState {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid state"})
	}
	delete(session.Values, "oauth_state")

	token, err := h.oidcCfg.OAuth2Config.Exchange(c.Request().Context(), c.QueryParam("code"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "token exchange failed"})
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "no id_token"})
	}

	idToken, err := h.oidcCfg.Verifier.Verify(c.Request().Context(), rawIDToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "token verification failed"})
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse claims"})
	}

	user, err := h.services.User.GetOrCreate(c.Request().Context(), claims.Sub, claims.Email, claims.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
	}

	session.Values[auth.SessionKeyUser] = user.ID
	session.Values[auth.SessionKeyEmail] = user.Email
	session.Values[auth.SessionKeyName] = user.Name
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save session"})
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) HandleLogout(c echo.Context) error {
	session, _ := h.store.Get(c.Request(), auth.SessionName)
	session.Options.MaxAge = -1
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to clear session"})
	}

	return c.Redirect(http.StatusFound, "/")
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
