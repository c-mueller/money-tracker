package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/service"
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

func TestAPITokenValidateExpired(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	plaintext, token, _ := svc.APIToken.Create(ctx, "Expiring Token")

	t.Run("expired token rejected", func(t *testing.T) {
		// Set ExpiresAt to the past via ent client
		pastTime := time.Now().Add(-1 * time.Hour)
		svc.client.APIToken.UpdateOneID(token.ID).SetExpiresAt(pastTime).SaveX(ctx)

		_, err := svc.APIToken.ValidateToken(ctx, plaintext)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden for expired token, got %v", err)
		}
	})

	t.Run("non-expired token accepted", func(t *testing.T) {
		// Set ExpiresAt to the future
		futureTime := time.Now().Add(24 * time.Hour)
		svc.client.APIToken.UpdateOneID(token.ID).SetExpiresAt(futureTime).SaveX(ctx)

		validatedToken, err := svc.APIToken.ValidateToken(ctx, plaintext)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if validatedToken.Name != "Expiring Token" {
			t.Errorf("Name = %q, want %q", validatedToken.Name, "Expiring Token")
		}
	})

	t.Run("token without expiry accepted", func(t *testing.T) {
		plain2, _, _ := svc.APIToken.Create(ctx, "No Expiry Token")
		validatedToken, err := svc.APIToken.ValidateToken(ctx, plain2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if validatedToken.Name != "No Expiry Token" {
			t.Errorf("Name = %q, want %q", validatedToken.Name, "No Expiry Token")
		}
	})
}

func TestAPITokenDelete(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	_, token, _ := svc.APIToken.Create(ctx, "To Delete")

	t.Run("success", func(t *testing.T) {
		if err := svc.APIToken.Delete(ctx, token.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		list, _ := svc.APIToken.List(ctx)
		if len(list) != 0 {
			t.Errorf("expected 0 tokens after delete, got %d", len(list))
		}
	})

	t.Run("other user cannot delete", func(t *testing.T) {
		_, token2, _ := svc.APIToken.Create(ctx, "User1 Token")

		user2, err := svc.User.GetOrCreate(t.Context(), "other-sub", "other@example.com", "Other User")
		if err != nil {
			t.Fatalf("failed to create user 2: %v", err)
		}
		ctx2 := service.WithUserID(t.Context(), user2.ID)

		err = svc.APIToken.Delete(ctx2, token2.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}

		// Verify token still exists for owner
		list, _ := svc.APIToken.List(ctx)
		found := false
		for _, tok := range list {
			if tok.ID == token2.ID {
				found = true
			}
		}
		if !found {
			t.Error("expected token to still exist for owner")
		}
	})

	t.Run("no auth", func(t *testing.T) {
		err := svc.APIToken.Delete(t.Context(), 1)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}
