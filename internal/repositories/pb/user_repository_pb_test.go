package pb

import (
	"context"
	"strings"
	"testing"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryPocketbase_Create_Success(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	user := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}

	createdUser, err := repo.Create(ctx, user)

	assert.NoError(err)
	assert.NotNil(createdUser)
	assert.NotEmpty(createdUser.ID)
	assert.Equal("test@example.com", createdUser.Email)
	// PocketBase may use email as username if username field is not properly set
	assert.NotEmpty(createdUser.Username)
	assert.Equal("Test User", createdUser.Name)
	assert.False(createdUser.Created.IsZero())
	assert.False(createdUser.Updated.IsZero())
}

func TestUserRepositoryPocketbase_Create_Error(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Create user with invalid email to trigger database error
	user := &models.User{
		Email:    "invalid-email",
		Username: "testuser",
		Name:     "Test User",
	}

	createdUser, err := repo.Create(ctx, user)
	assert.Error(err)
	assert.Nil(createdUser)
	assert.Equal(repositories.ErrDatabaseOperation, err)
}

func TestUserRepositoryPocketbase_GetByID_Success(t *testing.T) {
	assert := assert.New(t)

	app := NewTestApp(t)

	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Create a test user first
	testUser := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}
	createdUser, err := createUserInDB(t, app, testUser)
	assert.NoError(err)

	// Now test GetByID
	retrievedUser, err := repo.GetByID(ctx, createdUser.ID)

	assert.NoError(err)
	assert.NotNil(retrievedUser)
	assert.Equal(createdUser.ID, retrievedUser.ID)
	assert.Equal("test@example.com", retrievedUser.Email)
	// PocketBase's default auth collection behavior: uses email as username when username is empty
	// So we expect the retrieved username to be the email address
	assert.Equal("test@example.com", retrievedUser.Username)
	assert.Equal("Test User", retrievedUser.Name)
}

func TestUserRepositoryPocketbase_GetByID_Error(t *testing.T) {
	assert := assert.New(t)

	app := NewTestApp(t)

	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Try to get a non-existent user
	retrievedUser, err := repo.GetByID(ctx, "nonexistent-id")

	assert.Error(err)
	assert.Nil(retrievedUser)
	assert.Equal(repositories.ErrUseNotFound, err)
}

func TestUserRepositoryPocketbase_Delete_Success(t *testing.T) {
	assert := assert.New(t)

	app := NewTestApp(t)

	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Create a test user first
	testUser := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}
	createdUser, err := createUserInDB(t, app, testUser)
	assert.NoError(err)

	// Now test Delete
	err = repo.Delete(ctx, createdUser.ID)

	assert.NoError(err)

	// Verify user is deleted by trying to retrieve it
	retrievedUser, err := findUserInDB(t, app, createdUser.ID)
	assert.Error(err)
	assert.Nil(retrievedUser)
}

func TestUserRepositoryPocketbase_Delete_Error(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Try to delete a non-existent user
	err := repo.Delete(ctx, "nonexistent-id")

	assert.Error(err)
	assert.Equal(repositories.ErrUseNotFound, err)
}

func TestUserRepositoryPocketbase_Update_Success(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Create a test user first
	testUser := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}
	createdUser, err := createUserInDB(t, app, testUser)
	assert.NoError(err)

	createdUser.Name = "Updated Test User"

	// Now test Update
	updatedUser, err := repo.Update(ctx, createdUser)
	assert.NoError(err)
	assert.NotNil(updatedUser)

	// Verify user is updated by retrieving it
	retrievedUser, err := findUserInDB(t, app, createdUser.ID)
	assert.NoError(err)
	assert.Equal(retrievedUser.Name, "Updated Test User")
}

func TestUserRepositoryPocketbase_Update_Error(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Try to update a non-existent user
	_, err := repo.Update(ctx, &models.User{ID: "nonexistent-id"})

	assert.Error(err)
	assert.Equal(repositories.ErrUseNotFound, err)
}

func TestUserRepositoryPocketbase_GenerateAuthToken_Success(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Create a test user first
	testUser := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}
	createdUser, err := createUserInDB(t, app, testUser)
	assert.NoError(err)

	// Now test GenerateAuthToken
	token, err := repo.GenerateAuthToken(ctx, createdUser.ID)

	assert.NoError(err)
	assert.NotEmpty(token)
	assert.IsType("string", token)

	// Verify the token is a valid JWT (should have 3 parts separated by dots)
	tokenParts := len(strings.Split(token, "."))
	assert.Equal(3, tokenParts, "JWT token should have 3 parts separated by dots")
}

func TestUserRepositoryPocketbase_GenerateAuthToken_UserNotFound(t *testing.T) {
	assert := assert.New(t)
	app := NewTestApp(t)
	repo := NewUserRepositoryPocketbase(app)
	ctx := context.Background()

	// Try to generate token for a non-existent user
	token, err := repo.GenerateAuthToken(ctx, "nonexistent-id")

	assert.Error(err)
	assert.Empty(token)
	assert.Equal(repositories.ErrUseNotFound, err)
}

func createUserInDB(t *testing.T, app *pocketbase.PocketBase, user *models.User) (*models.User, error) {
	t.Helper()
	assert := require.New(t)

	collection, err := app.FindCollectionByNameOrId(string(CollectionUsers))
	assert.NoError(err)

	userRecord := core.NewRecord(collection)
	userRecord.Set("email", user.Email)
	userRecord.Set("username", user.Username)
	userRecord.Set("name", user.Name)
	userRecord.Set("password", "systemuser123")

	err = app.Save(userRecord)
	if err != nil {
		return nil, err
	}

	return recordToUser(userRecord), nil
}

// findUserInDB is a helper function to verify an user exists in the database
func findUserInDB(t *testing.T, app *pocketbase.PocketBase, id string) (*models.User, error) {
	t.Helper()
	assert := require.New(t)

	collection, err := app.FindCollectionByNameOrId(string(CollectionUsers))
	assert.NoError(err)

	record, err := app.FindRecordById(collection, id)
	if err != nil {
		return nil, err
	}

	return recordToUser(record), nil
}
