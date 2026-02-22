package service_test

import (
	"context"
	"errors"
	"testing"

	"icekalt.dev/money-tracker/internal/domain"
)

func TestUserGetByID(t *testing.T) {
	svc := setupTestServices(t)

	t.Run("success", func(t *testing.T) {
		user, err := svc.User.GetOrCreate(context.Background(), "get-by-id-sub", "get@example.com", "Get User")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := svc.User.GetByID(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != user.ID {
			t.Errorf("ID = %d, want %d", got.ID, user.ID)
		}
		if got.Email != "get@example.com" {
			t.Errorf("Email = %q, want %q", got.Email, "get@example.com")
		}
		if got.Name != "Get User" {
			t.Errorf("Name = %q, want %q", got.Name, "Get User")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.User.GetByID(context.Background(), 99999)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUserGetOrCreate(t *testing.T) {
	svc := setupTestServices(t)

	t.Run("creates new user", func(t *testing.T) {
		user, err := svc.User.GetOrCreate(context.Background(), "new-sub", "new@example.com", "New User")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Subject != "new-sub" {
			t.Errorf("Subject = %q, want %q", user.Subject, "new-sub")
		}
	})

	t.Run("returns existing user", func(t *testing.T) {
		user1, _ := svc.User.GetOrCreate(context.Background(), "existing-sub", "a@example.com", "First")
		user2, err := svc.User.GetOrCreate(context.Background(), "existing-sub", "b@example.com", "Second")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user1.ID != user2.ID {
			t.Errorf("expected same user ID, got %d and %d", user1.ID, user2.ID)
		}
	})
}
