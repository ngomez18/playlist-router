package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=user_service.go -destination=mocks/mock_user_service.go -package=mocks

type UserServicer interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

type UserService struct {
	userRepo repositories.UserRepository
	logger   *slog.Logger
}

func NewUserService(userRepo repositories.UserRepository, logger *slog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger.With("component", "UserService"),
	}
}

func (us *UserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	us.logger.InfoContext(ctx, "creating user", "email", user.Email, "name", user.Name)

	createdUser, err := us.userRepo.Create(ctx, user)
	if err != nil {
		us.logger.ErrorContext(ctx, "failed to create user", "error", err.Error())
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	us.logger.InfoContext(ctx, "user created successfully", "user_id", createdUser.ID, "email", createdUser.Email)

	return createdUser, nil
}

func (us *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	us.logger.InfoContext(ctx, "updating user", "user_id", user.ID, "email", user.Email)

	updatedUser, err := us.userRepo.Update(ctx, user)
	if err != nil {
		us.logger.ErrorContext(ctx, "failed to update user", "user_id", user.ID, "error", err.Error())
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	us.logger.InfoContext(ctx, "user updated successfully", "user_id", updatedUser.ID)

	return updatedUser, nil
}

func (us *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	us.logger.InfoContext(ctx, "retrieving user", "user_id", userID)

	user, err := us.userRepo.GetByID(ctx, userID)
	if err != nil {
		us.logger.ErrorContext(ctx, "failed to retrieve user", "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	us.logger.InfoContext(ctx, "user retrieved successfully", "user_id", user.ID)

	return user, nil
}

func (us *UserService) DeleteUser(ctx context.Context, userID string) error {
	us.logger.InfoContext(ctx, "deleting user", "user_id", userID)

	err := us.userRepo.Delete(ctx, userID)
	if err != nil {
		us.logger.ErrorContext(ctx, "failed to delete user", "user_id", userID, "error", err.Error())
		return fmt.Errorf("failed to delete user: %w", err)
	}

	us.logger.InfoContext(ctx, "user deleted successfully", "user_id", userID)

	return nil
}
