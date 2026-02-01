package services

import (
	"context"

	"money-buddy-backend/internal/models"
	"money-buddy-backend/internal/repositories"
)

type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
