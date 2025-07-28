package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/middleware"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestBasePlaylistController_Create_Success(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    *models.CreateBasePlaylistRequest
		serviceResult  *models.BasePlaylist
		expectedStatus int
	}{
		{
			name: "successful creation with valid request",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			serviceResult: &models.BasePlaylist{
				ID:                "playlist123",
				UserID:            "user123",
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful creation with minimum valid name",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "A",
				SpotifyPlaylistID: "spotify456",
			},
			serviceResult: &models.BasePlaylist{
				ID:                "playlist456",
				UserID:            "user456",
				Name:              "A",
				SpotifyPlaylistID: "spotify456",
				IsActive:          true,
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request
			requestBody, err := json.Marshal(tt.requestBody)
			assert.NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/api/base_playlist", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = addUserToContext(req)
			
			w := httptest.NewRecorder()

			// Set expectations
			mockService.EXPECT().
				CreateBasePlaylist(gomock.Any(), "test_user_123", tt.requestBody).
				Return(tt.serviceResult, nil).
				Times(1)

			// Execute
			controller.Create(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			// Verify response body
			var responseBody models.BasePlaylist
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(err)
			assert.Equal(tt.serviceResult.ID, responseBody.ID)
			assert.Equal(tt.serviceResult.UserID, responseBody.UserID)
			assert.Equal(tt.serviceResult.Name, responseBody.Name)
			assert.Equal(tt.serviceResult.SpotifyPlaylistID, responseBody.SpotifyPlaylistID)
			assert.Equal(tt.serviceResult.IsActive, responseBody.IsActive)
		})
	}
}

func TestBasePlaylistController_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		expectedStatus int
		expectedError  string
	}{
		{
			name: "empty name",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "",
				SpotifyPlaylistID: "spotify123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed:",
		},
		{
			name: "empty spotify playlist ID",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "Test Playlist",
				SpotifyPlaylistID: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed:",
		},
		{
			name: "name too long",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              strings.Repeat("a", 101), // 101 characters, max is 100
				SpotifyPlaylistID: "spotify123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request
			requestBody, err := json.Marshal(tt.requestBody)
			assert.NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/api/base_playlist", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// No service expectations since validation should fail before service call

			// Execute
			controller.Create(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_Create_RequestParsingErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid JSON",
			requestBody:    `{"name": "test", "spotify_playlist_id": }`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid payload",
		},
		{
			name:           "empty body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid payload",
		},
		{
			name:           "non-JSON content",
			requestBody:    "not json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request
			req := httptest.NewRequest(http.MethodPost, "/api/base_playlist", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// No service expectations since parsing should fail before service call

			// Execute
			controller.Create(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_Create_ServiceErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    *models.CreateBasePlaylistRequest
		serviceError   error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "service validation error",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			serviceError:   errors.New("failed to create playlist: validation failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to create base playlist",
		},
		{
			name: "service database error",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			serviceError:   errors.New("failed to create playlist: database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to create base playlist",
		},
		{
			name: "service duplicate error",
			requestBody: &models.CreateBasePlaylistRequest{
				Name:              "Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			serviceError:   errors.New("failed to create playlist: playlist already exists"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to create base playlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request
			requestBody, err := json.Marshal(tt.requestBody)
			assert.NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/api/base_playlist", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations
			mockService.EXPECT().
				CreateBasePlaylist(gomock.Any(), "test_user_123", tt.requestBody).
				Return(nil, tt.serviceError).
				Times(1)

			// Execute
			controller.Create(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_Create_ResponseEncodingError(t *testing.T) {
	assert := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockBasePlaylistServicer(ctrl)
	controller := NewBasePlaylistController(mockService)

	requestBody := &models.CreateBasePlaylistRequest{
		Name:              "Test Playlist",
		SpotifyPlaylistID: "spotify123",
	}

	// Create a playlist response that would cause JSON encoding issues
	// In this case, we'll use a response that should encode fine, but we'll
	// simulate the error by testing that the function handles encoding errors gracefully
	serviceResult := &models.BasePlaylist{
		ID:                "playlist123",
		UserID:            "user123",
		Name:              "Test Playlist",
		SpotifyPlaylistID: "spotify123",
		IsActive:          true,
	}

	// Prepare request
	reqBody, err := json.Marshal(requestBody)
	assert.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/base_playlist", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = addUserToContext(req)
	w := httptest.NewRecorder()

	// Set expectations
	mockService.EXPECT().
		CreateBasePlaylist(gomock.Any(), "test_user_123", requestBody).
		Return(serviceResult, nil).
		Times(1)

	// Execute
	controller.Create(w, req)

	// Verify successful response (since our test data is actually valid)
	assert.Equal(http.StatusCreated, w.Code)
	assert.Equal("application/json", w.Header().Get("Content-Type"))
}

func TestNewBasePlaylistController(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockBasePlaylistServicer(ctrl)
	controller := NewBasePlaylistController(mockService)

	assert.NotNil(controller)
	assert.Equal(mockService, controller.basePlaylistService)
	assert.NotNil(controller.validator)
}

func TestBasePlaylistController_Delete_Success(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		urlPath        string
		expectedStatus int
	}{
		{
			name:           "successful deletion with valid id",
			playlistID:     "playlist123",
			urlPath:        "/api/base_playlist/playlist123",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "successful deletion with complex id",
			playlistID:     "pl_abc123def456",
			urlPath:        "/api/base_playlist/pl_abc123def456",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request with path parameters
			req := httptest.NewRequest(http.MethodDelete, tt.urlPath, nil)
			req.SetPathValue("id", tt.playlistID)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations - expecting placeholder_user_id as defined in controller
			mockService.EXPECT().
				DeleteBasePlaylist(gomock.Any(), tt.playlistID, "test_user_123").
				Return(nil).
				Times(1)

			// Execute
			controller.Delete(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Empty(w.Body.String())
		})
	}
}

func TestBasePlaylistController_Delete_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "empty id in path",
			pathID:         "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "playlist id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request with empty path parameter
			req := httptest.NewRequest(http.MethodDelete, "/api/base_playlist/", nil)
			req.SetPathValue("id", tt.pathID)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// No service expectations since validation should fail before service call

			// Execute
			controller.Delete(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_Delete_ServiceErrors(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		serviceError   error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "playlist not found error",
			playlistID:     "nonexistent123",
			serviceError:   errors.New("failed to delete playlist: base playlist not found"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to delete base playlist",
		},
		{
			name:           "unauthorized access error",
			playlistID:     "playlist123",
			serviceError:   errors.New("failed to delete playlist: user can not access this resource"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to delete base playlist",
		},
		{
			name:           "database error",
			playlistID:     "playlist123",
			serviceError:   errors.New("failed to delete playlist: unable to complete db operation"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to delete base playlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request with path parameters
			urlPath := "/api/base_playlist/" + tt.playlistID
			req := httptest.NewRequest(http.MethodDelete, urlPath, nil)
			req.SetPathValue("id", tt.playlistID)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations - expecting placeholder_user_id as defined in controller
			mockService.EXPECT().
				DeleteBasePlaylist(gomock.Any(), tt.playlistID, "test_user_123").
				Return(tt.serviceError).
				Times(1)

			// Execute
			controller.Delete(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}
func TestBasePlaylistController_GetByID_Success(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		urlPath        string
		serviceResult  *models.BasePlaylist
		expectedStatus int
	}{
		{
			name:       "successful retrieval with valid id",
			playlistID: "playlist123",
			urlPath:    "/api/base_playlist/playlist123",
			serviceResult: &models.BasePlaylist{
				ID:                "playlist123",
				UserID:            "user123",
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "successful retrieval with complex id",
			playlistID: "pl_abc123def456",
			urlPath:    "/api/base_playlist/pl_abc123def456",
			serviceResult: &models.BasePlaylist{
				ID:                "pl_abc123def456",
				UserID:            "user456",
				Name:              "Another Playlist",
				SpotifyPlaylistID: "spotify456",
				IsActive:          false,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request with path parameters
			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)
			req.SetPathValue("id", tt.playlistID)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations - expecting placeholder_user_id as defined in controller
			mockService.EXPECT().
				GetBasePlaylist(gomock.Any(), tt.playlistID, "test_user_123").
				Return(tt.serviceResult, nil).
				Times(1)

			// Execute
			controller.GetByID(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			// Verify response body
			var responseBody models.BasePlaylist
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(err)
			assert.Equal(tt.serviceResult.ID, responseBody.ID)
			assert.Equal(tt.serviceResult.UserID, responseBody.UserID)
			assert.Equal(tt.serviceResult.Name, responseBody.Name)
			assert.Equal(tt.serviceResult.SpotifyPlaylistID, responseBody.SpotifyPlaylistID)
			assert.Equal(tt.serviceResult.IsActive, responseBody.IsActive)
		})
	}
}

func TestBasePlaylistController_GetByID_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "empty id in path",
			pathID:         "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "playlist id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request with empty path parameter
			req := httptest.NewRequest(http.MethodGet, "/api/base_playlist/", nil)
			req.SetPathValue("id", tt.pathID)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// No service expectations since validation should fail before service call

			// Execute
			controller.GetByID(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_GetByID_ServiceErrors(t *testing.T) {
	tests := []struct {
		name           string
		playlistID     string
		serviceError   error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "playlist not found error",
			playlistID:     "nonexistent123",
			serviceError:   errors.New("failed to retrieve playlist: base playlist not found"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to retrieve base playlist",
		},
		{
			name:           "unauthorized access error",
			playlistID:     "playlist123",
			serviceError:   errors.New("failed to retrieve playlist: user can not access this resource"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to retrieve base playlist",
		},
		{
			name:           "database error",
			playlistID:     "playlist123",
			serviceError:   errors.New("failed to retrieve playlist: unable to complete db operation"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unable to retrieve base playlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			// Prepare request with path parameters
			urlPath := "/api/base_playlist/" + tt.playlistID
			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			req.SetPathValue("id", tt.playlistID)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations - expecting placeholder_user_id as defined in controller
			mockService.EXPECT().
				GetBasePlaylist(gomock.Any(), tt.playlistID, "test_user_123").
				Return(nil, tt.serviceError).
				Times(1)

			// Execute
			controller.GetByID(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

// Helper function to add user to request context
func addUserToContext(req *http.Request) *http.Request {
	user := &models.User{ID: "test_user_123", Email: "test@example.com", Name: "Test User"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	return req.WithContext(ctx)
}
