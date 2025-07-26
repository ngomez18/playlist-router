package repositories

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=user_repository.go -destination=mocks/mock_user_repository.go -package=mocks

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
	Delete(ctx context.Context, userID string) error
}
