package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/orchestrators/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewSyncController(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrchestrator := mocks.NewMockSyncOrchestrator(ctrl)
	controller := NewSyncController(mockOrchestrator)

	assert.NotNil(controller)
	assert.Equal(mockOrchestrator, controller.syncOrchestrator)
}

func TestSyncController_SyncBasePlaylist_Success(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test data
	user := &models.User{ID: "user123"}
	basePlaylistID := "base456"
	expectedSyncEvent := &models.SyncEvent{
		ID:             "sync123",
		UserID:         user.ID,
		BasePlaylistID: basePlaylistID,
		Status:         models.SyncStatusInProgress,
	}

	// Setup mocks
	mockOrchestrator := mocks.NewMockSyncOrchestrator(ctrl)
	controller := NewSyncController(mockOrchestrator)

	mockOrchestrator.EXPECT().SyncBasePlaylist(gomock.Any(), user.ID, basePlaylistID).Return(expectedSyncEvent, nil)

	// Create request
	req := httptest.NewRequest("POST", "/api/base_playlist/"+basePlaylistID+"/sync", nil)
	req.SetPathValue("basePlaylistID", basePlaylistID)
	
	// Add user to context
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	req = req.WithContext(ctx)

	// Execute
	w := httptest.NewRecorder()
	controller.SyncBasePlaylist(w, req)

	// Assert
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("application/json", w.Header().Get("Content-Type"))
	assert.Contains(w.Body.String(), "sync123")
	assert.Contains(w.Body.String(), "user123")
	assert.Contains(w.Body.String(), "base456")
}

func TestSyncController_SyncBasePlaylist_NoUserInContext(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrchestrator := mocks.NewMockSyncOrchestrator(ctrl)
	controller := NewSyncController(mockOrchestrator)

	// Create request without user in context
	req := httptest.NewRequest("POST", "/api/base_playlist/base456/sync", nil)
	req.SetPathValue("basePlaylistID", "base456")

	w := httptest.NewRecorder()
	controller.SyncBasePlaylist(w, req)

	assert.Equal(http.StatusUnauthorized, w.Code)
	assert.Contains(w.Body.String(), "user not found in context")
}

func TestSyncController_SyncBasePlaylist_MissingBasePlaylistID(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrchestrator := mocks.NewMockSyncOrchestrator(ctrl)
	controller := NewSyncController(mockOrchestrator)

	user := &models.User{ID: "user123"}
	req := httptest.NewRequest("POST", "/api/base_playlist//sync", nil)
	// Don't set basePlaylistID path value to simulate missing ID
	
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	controller.SyncBasePlaylist(w, req)

	assert.Equal(http.StatusBadRequest, w.Code)
	assert.Contains(w.Body.String(), "base playlist ID is required")
}

func TestSyncController_SyncBasePlaylist_SyncInProgress(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := &models.User{ID: "user123"}
	basePlaylistID := "base456"

	mockOrchestrator := mocks.NewMockSyncOrchestrator(ctrl)
	controller := NewSyncController(mockOrchestrator)

	mockOrchestrator.EXPECT().SyncBasePlaylist(gomock.Any(), user.ID, basePlaylistID).Return(nil, errors.New("sync already in progress for base playlist "+basePlaylistID))

	req := httptest.NewRequest("POST", "/api/base_playlist/"+basePlaylistID+"/sync", nil)
	req.SetPathValue("basePlaylistID", basePlaylistID)
	
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	controller.SyncBasePlaylist(w, req)

	assert.Equal(http.StatusConflict, w.Code)
	assert.Contains(w.Body.String(), "sync already in progress")
}

func TestSyncController_SyncBasePlaylist_OrchestratorError(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := &models.User{ID: "user123"}
	basePlaylistID := "base456"

	mockOrchestrator := mocks.NewMockSyncOrchestrator(ctrl)
	controller := NewSyncController(mockOrchestrator)

	mockOrchestrator.EXPECT().SyncBasePlaylist(gomock.Any(), user.ID, basePlaylistID).Return(nil, errors.New("failed to aggregate track data"))

	req := httptest.NewRequest("POST", "/api/base_playlist/"+basePlaylistID+"/sync", nil)
	req.SetPathValue("basePlaylistID", basePlaylistID)
	
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	controller.SyncBasePlaylist(w, req)

	assert.Equal(http.StatusInternalServerError, w.Code)
	assert.Contains(w.Body.String(), "failed to sync base playlist")
}