package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"icekalt.dev/money-tracker/internal/domain"
)

type APITokenService struct {
	repo domain.APITokenRepo
}

func NewAPITokenService(repo domain.APITokenRepo) *APITokenService {
	return &APITokenService{repo: repo}
}

// Create generates a new API token and returns the plaintext token.
// The plaintext is only available at creation time.
func (s *APITokenService) Create(ctx context.Context, name string) (plaintext string, token *domain.APIToken, err error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return "", nil, fmt.Errorf("%w: no authenticated user", domain.ErrForbidden)
	}

	plain, err := generateToken()
	if err != nil {
		return "", nil, fmt.Errorf("generating token: %w", err)
	}

	hash := hashToken(plain)

	tok, err := s.repo.Create(ctx, &domain.APIToken{
		UserID:    userID,
		Name:      name,
		TokenHash: hash,
	})
	if err != nil {
		return "", nil, err
	}

	return plain, tok, nil
}

func (s *APITokenService) List(ctx context.Context) ([]*domain.APIToken, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%w: no authenticated user", domain.ErrForbidden)
	}

	return s.repo.ListByUser(ctx, userID)
}

func (s *APITokenService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *APITokenService) ValidateToken(ctx context.Context, plaintext string) (*domain.APIToken, error) {
	hash := hashToken(plaintext)
	return s.repo.GetByHash(ctx, hash)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "mt_" + hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
