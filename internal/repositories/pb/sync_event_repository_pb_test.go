package pb

import (
	"context"
	"testing"
	"time"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/pocketbase/pocketbase"
	"github.com/stretchr/testify/require"
)

func TestSyncEventRepositoryPocketbase_Create_Success(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		basePlaylistID   string
		childPlaylistIDs []string
		status           models.SyncStatus
		tracksProcessed  int
		totalAPIRequests int
		completedAt      *time.Time
		errorMessage     *string
	}{
		{
			name:             "successful creation with minimal data",
			userID:           "user123",
			basePlaylistID:   "base123",
			childPlaylistIDs: []string{},
			status:           models.SyncStatusInProgress,
			tracksProcessed:  0,
			totalAPIRequests: 0,
			completedAt:      nil,
			errorMessage:     nil,
		},
		{
			name:             "successful creation with child playlists",
			userID:           "user456",
			basePlaylistID:   "base456",
			childPlaylistIDs: []string{"child1", "child2", "child3"},
			status:           models.SyncStatusInProgress,
			tracksProcessed:  50,
			totalAPIRequests: 10,
			completedAt:      nil,
			errorMessage:     nil,
		},
		{
			name:             "successful creation with completed sync",
			userID:           "user789",
			basePlaylistID:   "base789",
			childPlaylistIDs: []string{"child4", "child5"},
			status:           models.SyncStatusCompleted,
			tracksProcessed:  100,
			totalAPIRequests: 25,
			completedAt:      ptrTime(time.Now()),
			errorMessage:     nil,
		},
		{
			name:             "successful creation with failed sync",
			userID:           "user999",
			basePlaylistID:   "base999",
			childPlaylistIDs: []string{"child6"},
			status:           models.SyncStatusFailed,
			tracksProcessed:  25,
			totalAPIRequests: 5,
			completedAt:      ptrTime(time.Now()),
			errorMessage:     ptrString("API rate limit exceeded"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupSyncEventCollection(t, app)
			repo := NewSyncEventRepositoryPocketbase(app)

			// Create sync event
			syncEvent := &models.SyncEvent{
				UserID:           tt.userID,
				BasePlaylistID:   tt.basePlaylistID,
				ChildPlaylistIDs: tt.childPlaylistIDs,
				Status:           tt.status,
				StartedAt:        time.Now(),
				TracksProcessed:  tt.tracksProcessed,
				TotalAPIRequests: tt.totalAPIRequests,
				CompletedAt:      tt.completedAt,
				ErrorMessage:     tt.errorMessage,
				Created:          time.Now(),
				Updated:          time.Now(),
			}

			// Execute test
			ctx := context.Background()
			createdSyncEvent, err := repo.Create(ctx, syncEvent)

			// Verify success
			assert.NoError(err)
			assert.NotNil(createdSyncEvent)

			// Verify sync event fields
			assert.Equal(tt.userID, createdSyncEvent.UserID)
			assert.Equal(tt.basePlaylistID, createdSyncEvent.BasePlaylistID)
			assert.Equal(tt.childPlaylistIDs, createdSyncEvent.ChildPlaylistIDs)
			assert.Equal(tt.status, createdSyncEvent.Status)
			assert.Equal(tt.tracksProcessed, createdSyncEvent.TracksProcessed)
			assert.Equal(tt.totalAPIRequests, createdSyncEvent.TotalAPIRequests)
			assert.NotEmpty(createdSyncEvent.ID)
			assert.NotZero(createdSyncEvent.StartedAt)
			assert.NotZero(createdSyncEvent.Created)
			assert.NotZero(createdSyncEvent.Updated)

			// Verify optional fields
			if tt.completedAt != nil {
				assert.NotNil(createdSyncEvent.CompletedAt)
				assert.WithinDuration(*tt.completedAt, *createdSyncEvent.CompletedAt, time.Second)
			} else {
				assert.Nil(createdSyncEvent.CompletedAt)
			}

			if tt.errorMessage != nil {
				assert.NotNil(createdSyncEvent.ErrorMessage)
				assert.Equal(*tt.errorMessage, *createdSyncEvent.ErrorMessage)
			} else {
				assert.Nil(createdSyncEvent.ErrorMessage)
			}

			// Verify the sync event was actually saved to the database
			savedSyncEvent, err := findSyncEventInDB(t, app, createdSyncEvent.ID)
			assert.NoError(err)
			assert.Equal(tt.userID, savedSyncEvent.UserID)
			assert.Equal(tt.basePlaylistID, savedSyncEvent.BasePlaylistID)
		})
	}
}

func TestSyncEventRepositoryPocketbase_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name            string
		userID          string
		basePlaylistID  string
		status          models.SyncStatus
		wantErrContains string
	}{
		{
			name:            "empty user ID",
			userID:          "",
			basePlaylistID:  "base123",
			status:          models.SyncStatusInProgress,
			wantErrContains: "user_id: cannot be blank",
		},
		{
			name:            "empty base playlist ID",
			userID:          "user123",
			basePlaylistID:  "",
			status:          models.SyncStatusInProgress,
			wantErrContains: "base_playlist_id: cannot be blank",
		},
		{
			name:            "empty status",
			userID:          "user123",
			basePlaylistID:  "base123",
			status:          "",
			wantErrContains: "status: cannot be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupSyncEventCollection(t, app)
			repo := NewSyncEventRepositoryPocketbase(app)

			// Create sync event with validation errors
			syncEvent := &models.SyncEvent{
				UserID:           tt.userID,
				BasePlaylistID:   tt.basePlaylistID,
				Status:           tt.status,
				StartedAt:        time.Now(),
				TracksProcessed:  0,
				TotalAPIRequests: 0,
				Created:          time.Now(),
				Updated:          time.Now(),
			}

			// Execute test
			ctx := context.Background()
			createdSyncEvent, err := repo.Create(ctx, syncEvent)

			// Verify error occurred
			assert.Error(err)
			assert.Nil(createdSyncEvent)
			assert.Contains(err.Error(), tt.wantErrContains)
		})
	}
}

func TestSyncEventRepositoryPocketbase_Update_Success(t *testing.T) {
	tests := []struct {
		name                   string
		initialStatus          models.SyncStatus
		initialTracksProcessed int
		initialAPIRequests     int
		updateStatus           models.SyncStatus
		updateTracksProcessed  int
		updateAPIRequests      int
		updateChildPlaylistIDs []string
		hasCompletedAt         bool
		errorMessage           *string
	}{
		{
			name:                   "successful completion update",
			initialStatus:          models.SyncStatusInProgress,
			initialTracksProcessed: 0,
			initialAPIRequests:     0,
			updateStatus:           models.SyncStatusCompleted,
			updateTracksProcessed:  150,
			updateAPIRequests:      30,
			updateChildPlaylistIDs: []string{"child1", "child2", "child3"},
			hasCompletedAt:         true,
			errorMessage:           nil,
		},
		{
			name:                   "failed sync with error message",
			initialStatus:          models.SyncStatusInProgress,
			initialTracksProcessed: 25,
			initialAPIRequests:     5,
			updateStatus:           models.SyncStatusFailed,
			updateTracksProcessed:  25,
			updateAPIRequests:      6,
			updateChildPlaylistIDs: []string{},
			hasCompletedAt:         true,
			errorMessage:           ptrString("Spotify API returned 429 - Rate limit exceeded"),
		},
		{
			name:                   "in-progress update with more tracks",
			initialStatus:          models.SyncStatusInProgress,
			initialTracksProcessed: 50,
			initialAPIRequests:     10,
			updateStatus:           models.SyncStatusInProgress,
			updateTracksProcessed:  100,
			updateAPIRequests:      20,
			updateChildPlaylistIDs: []string{"child1", "child2"},
			hasCompletedAt:         false,
			errorMessage:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupSyncEventCollection(t, app)
			repo := NewSyncEventRepositoryPocketbase(app)

			ctx := context.Background()

			// Create initial sync event
			initialSyncEvent := &models.SyncEvent{
				UserID:           "user123",
				BasePlaylistID:   "base123",
				ChildPlaylistIDs: []string{"child1", "child2"},
				Status:           tt.initialStatus,
				StartedAt:        time.Now(),
				TracksProcessed:  tt.initialTracksProcessed,
				TotalAPIRequests: tt.initialAPIRequests,
				Created:          time.Now(),
				Updated:          time.Now(),
			}

			createdSyncEvent, err := repo.Create(ctx, initialSyncEvent)
			assert.NoError(err)
			assert.NotNil(createdSyncEvent)

			// Prepare update
			var completedAt *time.Time
			if tt.hasCompletedAt {
				now := time.Now()
				completedAt = &now
			}

			updatedSyncEvent := &models.SyncEvent{
				Status:           tt.updateStatus,
				TracksProcessed:  tt.updateTracksProcessed,
				TotalAPIRequests: tt.updateAPIRequests,
				ChildPlaylistIDs: tt.updateChildPlaylistIDs,
				CompletedAt:      completedAt,
				ErrorMessage:     tt.errorMessage,
			}

			// Execute update
			result, err := repo.Update(ctx, createdSyncEvent.ID, updatedSyncEvent)
			assert.NoError(err)
			assert.NotNil(result)

			// Verify updated fields
			assert.Equal(tt.updateStatus, result.Status)
			assert.Equal(tt.updateTracksProcessed, result.TracksProcessed)
			assert.Equal(tt.updateAPIRequests, result.TotalAPIRequests)
			assert.Equal(tt.updateChildPlaylistIDs, result.ChildPlaylistIDs)

			// Verify optional fields
			if tt.hasCompletedAt {
				assert.NotNil(result.CompletedAt)
				assert.WithinDuration(*completedAt, *result.CompletedAt, time.Second)
			} else {
				assert.Nil(result.CompletedAt)
			}

			if tt.errorMessage != nil {
				assert.NotNil(result.ErrorMessage)
				assert.Equal(*tt.errorMessage, *result.ErrorMessage)
			} else {
				assert.Nil(result.ErrorMessage)
			}

			// Verify unchanged fields
			assert.Equal(createdSyncEvent.ID, result.ID)
			assert.Equal(createdSyncEvent.UserID, result.UserID)
			assert.Equal(createdSyncEvent.BasePlaylistID, result.BasePlaylistID)
			assert.WithinDuration(createdSyncEvent.StartedAt, result.StartedAt, time.Second)
		})
	}
}

func TestSyncEventRepositoryPocketbase_Update_NotFoundError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSyncEventCollection(t, app)
	repo := NewSyncEventRepositoryPocketbase(app)

	ctx := context.Background()

	// Try to update non-existent sync event
	updatedSyncEvent := &models.SyncEvent{
		Status:           models.SyncStatusCompleted,
		TracksProcessed:  100,
		TotalAPIRequests: 20,
	}

	result, err := repo.Update(ctx, "nonexistent123", updatedSyncEvent)

	// Verify error
	assert.Error(err)
	assert.Nil(result)
	assert.ErrorIs(err, repositories.ErrSyncEventNotFound)
}

func TestSyncEventRepositoryPocketbase_GetByID_Success(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSyncEventCollection(t, app)
	repo := NewSyncEventRepositoryPocketbase(app)

	ctx := context.Background()

	// Create sync event with all fields
	completedAt := time.Now()
	errorMessage := "Test error message"
	syncEvent := &models.SyncEvent{
		UserID:           "user123",
		BasePlaylistID:   "base123",
		ChildPlaylistIDs: []string{"child1", "child2"},
		Status:           models.SyncStatusFailed,
		StartedAt:        time.Now(),
		TracksProcessed:  75,
		TotalAPIRequests: 15,
		CompletedAt:      &completedAt,
		ErrorMessage:     &errorMessage,
		Created:          time.Now(),
		Updated:          time.Now(),
	}

	createdSyncEvent, err := repo.Create(ctx, syncEvent)
	assert.NoError(err)

	// Execute GetByID
	retrievedSyncEvent, err := repo.GetByID(ctx, createdSyncEvent.ID)
	assert.NoError(err)
	assert.NotNil(retrievedSyncEvent)

	// Verify the retrieved sync event matches the created one
	assert.Equal(createdSyncEvent.ID, retrievedSyncEvent.ID)
	assert.Equal(createdSyncEvent.UserID, retrievedSyncEvent.UserID)
	assert.Equal(createdSyncEvent.BasePlaylistID, retrievedSyncEvent.BasePlaylistID)
	assert.Equal(createdSyncEvent.ChildPlaylistIDs, retrievedSyncEvent.ChildPlaylistIDs)
	assert.Equal(createdSyncEvent.Status, retrievedSyncEvent.Status)
	assert.Equal(createdSyncEvent.TracksProcessed, retrievedSyncEvent.TracksProcessed)
	assert.Equal(createdSyncEvent.TotalAPIRequests, retrievedSyncEvent.TotalAPIRequests)
	assert.WithinDuration(createdSyncEvent.StartedAt, retrievedSyncEvent.StartedAt, time.Second)

	// Verify optional fields
	assert.NotNil(retrievedSyncEvent.CompletedAt)
	assert.WithinDuration(*createdSyncEvent.CompletedAt, *retrievedSyncEvent.CompletedAt, time.Second)
	assert.NotNil(retrievedSyncEvent.ErrorMessage)
	assert.Equal(*createdSyncEvent.ErrorMessage, *retrievedSyncEvent.ErrorMessage)
}

func TestSyncEventRepositoryPocketbase_GetByID_NotFoundError(t *testing.T) {
	assert := require.New(t)

	// Setup test environment
	app := NewTestApp(t)
	SetupSyncEventCollection(t, app)
	repo := NewSyncEventRepositoryPocketbase(app)

	ctx := context.Background()

	// Execute GetByID with non-existent ID
	retrievedSyncEvent, err := repo.GetByID(ctx, "nonexistent123")

	// Verify error
	assert.Error(err)
	assert.Nil(retrievedSyncEvent)
	assert.ErrorIs(err, repositories.ErrSyncEventNotFound)
}

func TestSyncEventRepositoryPocketbase_GetByUserID_Success(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		syncEventsToCreate []struct {
			basePlaylistID string
			status         models.SyncStatus
		}
		expectedCount int
	}{
		{
			name:   "user with multiple sync events",
			userID: "user123",
			syncEventsToCreate: []struct {
				basePlaylistID string
				status         models.SyncStatus
			}{
				{"base1", models.SyncStatusCompleted},
				{"base2", models.SyncStatusInProgress},
				{"base3", models.SyncStatusFailed},
			},
			expectedCount: 3,
		},
		{
			name:   "user with single sync event",
			userID: "user456",
			syncEventsToCreate: []struct {
				basePlaylistID string
				status         models.SyncStatus
			}{
				{"base4", models.SyncStatusInProgress},
			},
			expectedCount: 1,
		},
		{
			name:   "user with no sync events",
			userID: "user789",
			syncEventsToCreate: []struct {
				basePlaylistID string
				status         models.SyncStatus
			}{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupSyncEventCollection(t, app)
			repo := NewSyncEventRepositoryPocketbase(app)

			ctx := context.Background()

			// Create sync events for this user
			createdSyncEvents := make([]*models.SyncEvent, 0, len(tt.syncEventsToCreate))
			for _, syncData := range tt.syncEventsToCreate {
				syncEvent := &models.SyncEvent{
					UserID:           tt.userID,
					BasePlaylistID:   syncData.basePlaylistID,
					Status:           syncData.status,
					StartedAt:        time.Now(),
					TracksProcessed:  0,
					TotalAPIRequests: 0,
					Created:          time.Now(),
					Updated:          time.Now(),
				}
				created, err := repo.Create(ctx, syncEvent)
				assert.NoError(err)
				createdSyncEvents = append(createdSyncEvents, created)
			}

			// Create sync events for different users to ensure isolation
			otherUserSyncEvent := &models.SyncEvent{
				UserID:           "other_user",
				BasePlaylistID:   "other_base",
				Status:           models.SyncStatusCompleted,
				StartedAt:        time.Now(),
				TracksProcessed:  0,
				TotalAPIRequests: 0,
				Created:          time.Now(),
				Updated:          time.Now(),
			}
			_, err := repo.Create(ctx, otherUserSyncEvent)
			assert.NoError(err)

			// Execute GetByUserID
			retrievedSyncEvents, err := repo.GetByUserID(ctx, tt.userID)

			// Verify success
			assert.NoError(err)
			assert.NotNil(retrievedSyncEvents)
			assert.Len(retrievedSyncEvents, tt.expectedCount)

			// If we have sync events, verify they match what we created
			if tt.expectedCount > 0 {
				// Verify all retrieved sync events belong to the correct user
				for _, syncEvent := range retrievedSyncEvents {
					assert.Equal(tt.userID, syncEvent.UserID)
					assert.NotEmpty(syncEvent.ID)
					assert.NotEmpty(syncEvent.BasePlaylistID)
				}

				// Verify specific sync event data matches
				basePlaylistIDs := make(map[string]bool)
				for _, syncEvent := range retrievedSyncEvents {
					basePlaylistIDs[syncEvent.BasePlaylistID] = true
				}

				// Verify all created sync events are present
				for _, created := range createdSyncEvents {
					assert.True(basePlaylistIDs[created.BasePlaylistID], "Sync event for base playlist %s should be in results", created.BasePlaylistID)
				}
			}
		})
	}
}

func TestSyncEventRepositoryPocketbase_GetByBasePlaylistID_Success(t *testing.T) {
	tests := []struct {
		name               string
		basePlaylistID     string
		syncEventsToCreate []struct {
			userID string
			status models.SyncStatus
		}
		expectedCount int
	}{
		{
			name:           "base playlist with multiple sync events",
			basePlaylistID: "base123",
			syncEventsToCreate: []struct {
				userID string
				status models.SyncStatus
			}{
				{"user1", models.SyncStatusCompleted},
				{"user1", models.SyncStatusInProgress},
				{"user2", models.SyncStatusFailed},
			},
			expectedCount: 3,
		},
		{
			name:           "base playlist with single sync event",
			basePlaylistID: "base456",
			syncEventsToCreate: []struct {
				userID string
				status models.SyncStatus
			}{
				{"user3", models.SyncStatusCompleted},
			},
			expectedCount: 1,
		},
		{
			name:           "base playlist with no sync events",
			basePlaylistID: "base789",
			syncEventsToCreate: []struct {
				userID string
				status models.SyncStatus
			}{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup test environment
			app := NewTestApp(t)
			SetupSyncEventCollection(t, app)
			repo := NewSyncEventRepositoryPocketbase(app)

			ctx := context.Background()

			// Create sync events for this base playlist
			createdSyncEvents := make([]*models.SyncEvent, 0, len(tt.syncEventsToCreate))
			for _, syncData := range tt.syncEventsToCreate {
				syncEvent := &models.SyncEvent{
					UserID:           syncData.userID,
					BasePlaylistID:   tt.basePlaylistID,
					Status:           syncData.status,
					StartedAt:        time.Now(),
					TracksProcessed:  0,
					TotalAPIRequests: 0,
					Created:          time.Now(),
					Updated:          time.Now(),
				}
				created, err := repo.Create(ctx, syncEvent)
				assert.NoError(err)
				createdSyncEvents = append(createdSyncEvents, created)
			}

			// Create sync events for different base playlists to ensure isolation
			otherBasePlaylistSyncEvent := &models.SyncEvent{
				UserID:           "user999",
				BasePlaylistID:   "other_base",
				Status:           models.SyncStatusCompleted,
				StartedAt:        time.Now(),
				TracksProcessed:  0,
				TotalAPIRequests: 0,
				Created:          time.Now(),
				Updated:          time.Now(),
			}
			_, err := repo.Create(ctx, otherBasePlaylistSyncEvent)
			assert.NoError(err)

			// Execute GetByBasePlaylistID
			retrievedSyncEvents, err := repo.GetByBasePlaylistID(ctx, tt.basePlaylistID)

			// Verify success
			assert.NoError(err)
			assert.NotNil(retrievedSyncEvents)
			assert.Len(retrievedSyncEvents, tt.expectedCount)

			// If we have sync events, verify they match what we created
			if tt.expectedCount > 0 {
				// Verify all retrieved sync events belong to the correct base playlist
				for _, syncEvent := range retrievedSyncEvents {
					assert.Equal(tt.basePlaylistID, syncEvent.BasePlaylistID)
					assert.NotEmpty(syncEvent.ID)
					assert.NotEmpty(syncEvent.UserID)
				}

				// Verify specific sync event data matches
				userIDs := make(map[string]int)
				for _, syncEvent := range retrievedSyncEvents {
					userIDs[syncEvent.UserID]++
				}

				// Verify all created sync events are present
				expectedUserCounts := make(map[string]int)
				for _, created := range createdSyncEvents {
					expectedUserCounts[created.UserID]++
				}

				assert.Equal(expectedUserCounts, userIDs, "Retrieved sync events should match created sync events by user")
			}
		})
	}
}

// Helper functions

// ptrTime returns a pointer to a time.Time value
func ptrTime(t time.Time) *time.Time {
	return &t
}

// ptrString returns a pointer to a string value
func ptrString(s string) *string {
	return &s
}

// findSyncEventInDB is a helper function to verify a sync event exists in the database
func findSyncEventInDB(t *testing.T, app *pocketbase.PocketBase, id string) (*models.SyncEvent, error) {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId(string(CollectionSyncEvent))
	if err != nil {
		return nil, err
	}

	record, err := app.FindRecordById(collection, id)
	if err != nil {
		return nil, err
	}

	return recordToSyncEvent(record), nil
}
