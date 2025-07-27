package pb

import (
	"context"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type UserRepositoryPocketbase struct {
	collection Collection
	app        *pocketbase.PocketBase
	log        *slog.Logger
}

func NewUserRepositoryPocketbase(pb *pocketbase.PocketBase) *UserRepositoryPocketbase {
	return &UserRepositoryPocketbase{
		app:        pb,
		collection: CollectionUsers,
		log:        pb.Logger().With("component", "UserRepository"),
	}
}

func (uRepo *UserRepositoryPocketbase) Create(ctx context.Context, user *models.User) (*models.User, error) {
	collection, err := GetCollection(ctx, uRepo.app, uRepo.collection)
	if err != nil {
		return nil, err
	}

	userRecord := core.NewRecord(collection)
	userRecord.Set("email", user.Email)
	userRecord.Set("username", user.Username)
	userRecord.Set("name", user.Name)
	userRecord.Set("password", "systemuser123")
	userRecord.Set("passwordConfirm", "systemuser123")

	if err := uRepo.app.Save(userRecord); err != nil {
		uRepo.log.ErrorContext(ctx, "unable to store user record", "record", userRecord, "error", err)
		return nil, repositories.ErrDatabaseOperation
	}

	createdUser := recordToUser(userRecord)
	uRepo.log.InfoContext(ctx, "user created successfully", "user", createdUser)

	return createdUser, nil
}

func (uRepo *UserRepositoryPocketbase) Update(ctx context.Context, user *models.User) (*models.User, error) {
	userRecord, err := uRepo.app.FindRecordById(string(uRepo.collection), user.ID)
	if err != nil {
		uRepo.log.ErrorContext(ctx, "unable to fetch user", "user", user.ID, "error", err)
		return nil, repositories.ErrUseNotFound
	}

	userRecord.Set("email", user.Email)
	userRecord.Set("username", user.Username)
	userRecord.Set("name", user.Name)

	if err := uRepo.app.Save(userRecord); err != nil {
		uRepo.log.ErrorContext(ctx, "unable to store user record", "record", userRecord, "error", err)
		return nil, repositories.ErrDatabaseOperation
	}

	createdUser := recordToUser(userRecord)
	uRepo.log.InfoContext(ctx, "user updated successfully", "user", createdUser)

	return createdUser, nil
}

func (uRepo *UserRepositoryPocketbase) GetByID(ctx context.Context, userID string) (*models.User, error) {
	record, err := uRepo.app.FindRecordById(string(uRepo.collection), userID)
	if err != nil {
		uRepo.log.ErrorContext(ctx, "unable to fetch user", "user", userID, "error", err)
		return nil, repositories.ErrUseNotFound
	}

	user := recordToUser(record)
	uRepo.log.InfoContext(ctx, "user retrieved successfully", "user", user)

	return user, nil
}

func (uRepo *UserRepositoryPocketbase) Delete(ctx context.Context, userID string) error {
	record, err := uRepo.app.FindRecordById(string(uRepo.collection), userID)
	if err != nil {
		uRepo.log.ErrorContext(ctx, "unable to fetch user", "user", userID, "error", err)
		return repositories.ErrUseNotFound
	}

	if err := uRepo.app.Delete(record); err != nil {
		uRepo.log.ErrorContext(ctx, "unable to delete user record", "user", userID, "error", err)
		return repositories.ErrDatabaseOperation
	}

	uRepo.log.InfoContext(ctx, "user deleted successfully", "user", userID)
	return nil
}

// recordToUser converts a PocketBase record to a User model
// Note: PocketBase's default auth collection may use email as the username field
// if the username field is not properly configured or populated
func recordToUser(record *core.Record) *models.User {
	// Try different possible field names for username
	username := record.GetString("username")
	if username == "" {
		username = record.GetString("email") // Fallback to email if username is empty
	}

	return &models.User{
		ID:       record.Id,
		Created:  record.GetDateTime("created").Time(),
		Username: username,
		Email:    record.GetString("email"),
		Name:     record.GetString("name"),
		Updated:  record.GetDateTime("updated").Time(),
	}
}
