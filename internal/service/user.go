package service

import (
	"context"

	"icekalt.dev/money-tracker/internal/domain"
)

type UserService struct {
	repo domain.UserRepo
}

func NewUserService(repo domain.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetOrCreate(ctx context.Context, subject, email, name string) (*domain.User, error) {
	user, err := s.repo.GetBySubject(ctx, subject)
	if err == nil {
		return user, nil
	}

	return s.repo.Create(ctx, &domain.User{
		Email:   email,
		Name:    name,
		Subject: subject,
	})
}

func (s *UserService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
