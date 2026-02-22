package repository

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/ent"
	entuser "icekalt.dev/money-tracker/ent/user"
	"icekalt.dev/money-tracker/internal/domain"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	u, err := r.client.User.Create().
		SetEmail(user.Email).
		SetName(user.Name).
		SetSubject(user.Subject).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("%w: user already exists", domain.ErrConflict)
		}
		return nil, err
	}
	return userToDomain(u), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	u, err := r.client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: user %d", domain.ErrNotFound, id)
		}
		return nil, err
	}
	return userToDomain(u), nil
}

func (r *UserRepository) GetBySubject(ctx context.Context, subject string) (*domain.User, error) {
	u, err := r.client.User.Query().
		Where(entuser.SubjectEQ(subject)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: user with subject %s", domain.ErrNotFound, subject)
		}
		return nil, err
	}
	return userToDomain(u), nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	u, err := r.client.User.UpdateOneID(user.ID).
		SetEmail(user.Email).
		SetName(user.Name).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: user %d", domain.ErrNotFound, user.ID)
		}
		return nil, err
	}
	return userToDomain(u), nil
}
