package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewSyncEventService(t *testing.T) {
	require := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	service := NewSyncEventService(mockRepo, logger)

	require.NotNil(service)
	require.Equal(mockRepo, service.syncEventRepo)
	require.NotNil(service.logger)
}

func TestSyncEventService_CreateSyncEvent_Success(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	now := time.Now()

	inputSyncEvent := &models.SyncEvent{
		UserID:           "user123",
		BasePlaylistID:   "base123",
		ChildPlaylistIDs: []string{"child1", "child2"},
		Status:           models.SyncStatusInProgress,
		StartedAt:        now,
		TracksProcessed:  0,
		TotalAPIRequests: 0,
		Created:          now,
		Updated:          now,
	}

	expectedSyncEvent := &models.SyncEvent{
		ID:               "sync123",
		UserID:           "user123",
		BasePlaylistID:   "base123",
		ChildPlaylistIDs: []string{"child1", "child2"},
		Status:           models.SyncStatusInProgress,
		StartedAt:        now,
		TracksProcessed:  0,
		TotalAPIRequests: 0,
		Created:          now,
		Updated:          now,
	}

	// Set expectations
	mockRepo.EXPECT().
		Create(ctx, inputSyncEvent).
		Return(expectedSyncEvent, nil).
		Times(1)

	// Execute
	result, err := service.CreateSyncEvent(ctx, inputSyncEvent)

	// Verify
	require.NoError(err)
	require.NotNil(result)
	require.Equal(expectedSyncEvent, result)
}

func TestSyncEventService_CreateSyncEvent_Error(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	now := time.Now()

	inputSyncEvent := &models.SyncEvent{
		UserID:           "user123",
		BasePlaylistID:   "base123",
		Status:           models.SyncStatusInProgress,
		StartedAt:        now,
		TracksProcessed:  0,
		TotalAPIRequests: 0,
		Created:          now,
		Updated:          now,
	}

	// Set expectations
	mockRepo.EXPECT().
		Create(ctx, inputSyncEvent).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	result, err := service.CreateSyncEvent(ctx, inputSyncEvent)

	// Verify
	require.Error(err)
	require.Nil(result)
	require.Contains(err.Error(), "failed to create sync event")
	require.ErrorIs(err, repositories.ErrDatabaseOperation)
}

func TestSyncEventService_UpdateSyncEvent_Success(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	syncID := "sync123"
	now := time.Now()

	updateSyncEvent := &models.SyncEvent{
		Status:           models.SyncStatusCompleted,
		TracksProcessed:  100,
		TotalAPIRequests: 25,
		CompletedAt:      &now,
	}

	expectedSyncEvent := &models.SyncEvent{
		ID:               syncID,
		UserID:           "user123",
		BasePlaylistID:   "base123",
		Status:           models.SyncStatusCompleted,
		TracksProcessed:  100,
		TotalAPIRequests: 25,
		CompletedAt:      &now,
	}

	// Set expectations
	mockRepo.EXPECT().
		Update(ctx, syncID, updateSyncEvent).
		Return(expectedSyncEvent, nil).
		Times(1)

	// Execute
	result, err := service.UpdateSyncEvent(ctx, syncID, updateSyncEvent)

	// Verify
	require.NoError(err)
	require.NotNil(result)
	require.Equal(expectedSyncEvent, result)
}

func TestSyncEventService_UpdateSyncEvent_Error(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	syncID := "nonexistent"

	updateSyncEvent := &models.SyncEvent{
		Status: models.SyncStatusCompleted,
	}

	// Set expectations
	mockRepo.EXPECT().
		Update(ctx, syncID, updateSyncEvent).
		Return(nil, repositories.ErrSyncEventNotFound).
		Times(1)

	// Execute
	result, err := service.UpdateSyncEvent(ctx, syncID, updateSyncEvent)

	// Verify
	require.Error(err)
	require.Nil(result)
	require.Contains(err.Error(), "failed to update sync event")
	require.ErrorIs(err, repositories.ErrSyncEventNotFound)
}

func TestSyncEventService_GetSyncEvent_Success(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	syncID := "sync123"

	expectedSyncEvent := &models.SyncEvent{
		ID:               syncID,
		UserID:           "user123",
		BasePlaylistID:   "base123",
		Status:           models.SyncStatusCompleted,
		TracksProcessed:  100,
		TotalAPIRequests: 25,
	}

	// Set expectations
	mockRepo.EXPECT().
		GetByID(ctx, syncID).
		Return(expectedSyncEvent, nil).
		Times(1)

	// Execute
	result, err := service.GetSyncEvent(ctx, syncID)

	// Verify
	require.NoError(err)
	require.NotNil(result)
	require.Equal(expectedSyncEvent, result)
}

func TestSyncEventService_GetSyncEvent_Error(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	syncID := "nonexistent"

	// Set expectations
	mockRepo.EXPECT().
		GetByID(ctx, syncID).
		Return(nil, repositories.ErrSyncEventNotFound).
		Times(1)

	// Execute
	result, err := service.GetSyncEvent(ctx, syncID)

	// Verify
	require.Error(err)
	require.Nil(result)
	require.Contains(err.Error(), "failed to retrieve sync event")
	require.ErrorIs(err, repositories.ErrSyncEventNotFound)
}

func TestSyncEventService_HasActiveSyncForBasePlaylist_Success(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		basePlaylistID string
		syncEvents     []*models.SyncEvent
		expected       bool
	}{
		{
			name:           "has active sync for user and base playlist",
			userID:         "user123",
			basePlaylistID: "base123",
			syncEvents: []*models.SyncEvent{
				{
					ID:             "sync1",
					UserID:         "user123",
					BasePlaylistID: "base123",
					Status:         models.SyncStatusInProgress,
				},
				{
					ID:             "sync2",
					UserID:         "user123",
					BasePlaylistID: "base123",
					Status:         models.SyncStatusCompleted,
				},
			},
			expected: true,
		},
		{
			name:           "no active sync - different user",
			userID:         "user123",
			basePlaylistID: "base123",
			syncEvents: []*models.SyncEvent{
				{
					ID:             "sync1",
					UserID:         "user456",
					BasePlaylistID: "base123",
					Status:         models.SyncStatusInProgress,
				},
			},
			expected: false,
		},
		{
			name:           "no active sync - all completed",
			userID:         "user123",
			basePlaylistID: "base123",
			syncEvents: []*models.SyncEvent{
				{
					ID:             "sync1",
					UserID:         "user123",
					BasePlaylistID: "base123",
					Status:         models.SyncStatusCompleted,
				},
				{
					ID:             "sync2",
					UserID:         "user123",
					BasePlaylistID: "base123",
					Status:         models.SyncStatusFailed,
				},
			},
			expected: false,
		},
		{
			name:           "no sync events",
			userID:         "user123",
			basePlaylistID: "base123",
			syncEvents:     []*models.SyncEvent{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSyncEventRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewSyncEventService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByBasePlaylistID(ctx, tt.basePlaylistID).
				Return(tt.syncEvents, nil).
				Times(1)

			// Execute
			result, err := service.HasActiveSyncForBasePlaylist(ctx, tt.userID, tt.basePlaylistID)

			// Verify
			require.NoError(err)
			require.Equal(tt.expected, result)
		})
	}
}

func TestSyncEventService_HasActiveSyncForBasePlaylist_Error(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	userID := "user123"
	basePlaylistID := "base123"

	// Set expectations
	mockRepo.EXPECT().
		GetByBasePlaylistID(ctx, basePlaylistID).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	result, err := service.HasActiveSyncForBasePlaylist(ctx, userID, basePlaylistID)

	// Verify
	require.Error(err)
	require.False(result)
	require.Contains(err.Error(), "failed to check for active sync")
	require.ErrorIs(err, repositories.ErrDatabaseOperation)
}

func TestSyncEventService_HasActiveSyncForUser_Success(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		syncEvents []*models.SyncEvent
		expected   bool
	}{
		{
			name:   "has active sync for user",
			userID: "user123",
			syncEvents: []*models.SyncEvent{
				{
					ID:     "sync1",
					UserID: "user123",
					Status: models.SyncStatusInProgress,
				},
				{
					ID:     "sync2",
					UserID: "user123",
					Status: models.SyncStatusCompleted,
				},
			},
			expected: true,
		},
		{
			name:   "no active sync - all completed/failed",
			userID: "user123",
			syncEvents: []*models.SyncEvent{
				{
					ID:     "sync1",
					UserID: "user123",
					Status: models.SyncStatusCompleted,
				},
				{
					ID:     "sync2",
					UserID: "user123",
					Status: models.SyncStatusFailed,
				},
			},
			expected: false,
		},
		{
			name:       "no sync events",
			userID:     "user123",
			syncEvents: []*models.SyncEvent{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSyncEventRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewSyncEventService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByUserID(ctx, tt.userID).
				Return(tt.syncEvents, nil).
				Times(1)

			// Execute
			result, err := service.HasActiveSyncForUser(ctx, tt.userID)

			// Verify
			require.NoError(err)
			require.Equal(tt.expected, result)
		})
	}
}

func TestSyncEventService_HasActiveSyncForUser_Error(t *testing.T) {
	require := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSyncEventRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewSyncEventService(mockRepo, logger)

	ctx := context.Background()
	userID := "user123"

	// Set expectations
	mockRepo.EXPECT().
		GetByUserID(ctx, userID).
		Return(nil, errors.New("database error")).
		Times(1)

	// Execute
	result, err := service.HasActiveSyncForUser(ctx, userID)

	// Verify
	require.Error(err)
	require.False(result)
	require.Contains(err.Error(), "failed to check for active sync")
}
