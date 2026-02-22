package service_test

import (
	"errors"
	"strings"
	"testing"

	"icekalt.dev/money-tracker/internal/domain"
)

func TestAPITokenCreate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	t.Run("success", func(t *testing.T) {
		plaintext, token, err := svc.APIToken.Create(ctx, "My Token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(plaintext, "mt_") {
			t.Errorf("plaintext should start with mt_, got %q", plaintext)
		}
		if token.Name != "My Token" {
			t.Errorf("Name = %q, want %q", token.Name, "My Token")
		}
		if token.TokenHash == "" {
			t.Error("expected TokenHash to be set")
		}
		if token.TokenHash == plaintext {
			t.Error("TokenHash should not equal plaintext")
		}
	})

	t.Run("no auth", func(t *testing.T) {
		_, _, err := svc.APIToken.Create(t.Context(), "Test")
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestAPITokenValidate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	plaintext, _, _ := svc.APIToken.Create(ctx, "Validate Token")

	t.Run("valid token", func(t *testing.T) {
		token, err := svc.APIToken.ValidateToken(ctx, plaintext)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token.Name != "Validate Token" {
			t.Errorf("Name = %q, want %q", token.Name, "Validate Token")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := svc.APIToken.ValidateToken(ctx, "mt_invalid_token_hash")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAPITokenList(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	svc.APIToken.Create(ctx, "Token 1")
	svc.APIToken.Create(ctx, "Token 2")

	list, err := svc.APIToken.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(list))
	}
}

func TestAPITokenDelete(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	_, token, _ := svc.APIToken.Create(ctx, "To Delete")

	if err := svc.APIToken.Delete(ctx, token.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list, _ := svc.APIToken.List(ctx)
	if len(list) != 0 {
		t.Errorf("expected 0 tokens after delete, got %d", len(list))
	}
}
