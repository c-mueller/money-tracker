package repository

import (
	"context"
	"fmt"
	"time"

	"icekalt.dev/money-tracker/ent"
	entapitoken "icekalt.dev/money-tracker/ent/apitoken"
	entuser "icekalt.dev/money-tracker/ent/user"
	"icekalt.dev/money-tracker/internal/domain"
)

type APITokenRepository struct {
	client *ent.Client
}

func NewAPITokenRepository(client *ent.Client) *APITokenRepository {
	return &APITokenRepository{client: client}
}

func (r *APITokenRepository) Create(ctx context.Context, token *domain.APIToken) (*domain.APIToken, error) {
	q := r.client.APIToken.Create().
		SetName(token.Name).
		SetTokenHash(token.TokenHash).
		SetUserID(token.UserID)

	if token.ExpiresAt != nil {
		q.SetExpiresAt(*token.ExpiresAt)
	}

	t, err := q.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("%w: token hash already exists", domain.ErrConflict)
		}
		return nil, err
	}
	t.Edges.User = &ent.User{ID: token.UserID}
	return apiTokenToDomain(t), nil
}

func (r *APITokenRepository) GetByHash(ctx context.Context, hash string) (*domain.APIToken, error) {
	t, err := r.client.APIToken.Query().
		Where(entapitoken.TokenHashEQ(hash)).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: api token", domain.ErrNotFound)
		}
		return nil, err
	}
	return apiTokenToDomain(t), nil
}

func (r *APITokenRepository) ListByUser(ctx context.Context, userID int) ([]*domain.APIToken, error) {
	items, err := r.client.APIToken.Query().
		Where(entapitoken.HasUserWith(entuser.IDEQ(userID))).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.APIToken, 0, len(items))
	for _, t := range items {
		result = append(result, apiTokenToDomain(t))
	}
	return result, nil
}

func (r *APITokenRepository) UpdateLastUsed(ctx context.Context, id int, t time.Time) error {
	return r.client.APIToken.UpdateOneID(id).SetLastUsed(t).Exec(ctx)
}

func (r *APITokenRepository) Delete(ctx context.Context, id int) error {
	err := r.client.APIToken.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: api token %d", domain.ErrNotFound, id)
		}
		return err
	}
	return nil
}
