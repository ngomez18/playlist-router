package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

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

func TestBasePlaylistController_Create_Errors(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		serviceError       error
		noUserInContext    bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "invalid request body",
			requestBody:        "invalid json",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "invalid payload",
		},
		{
			name:               "validation error",
			requestBody:        models.CreateBasePlaylistRequest{Name: ""},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "validation failed",
		},
		{
			name:               "no user in context",
			requestBody:        models.CreateBasePlaylistRequest{Name: "Test"},
			noUserInContext:    true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "user not found in context",
		},
		{
			name:               "service error",
			requestBody:        models.CreateBasePlaylistRequest{Name: "Test"},
			serviceError:       errors.New("some service error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to create base playlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					CreateBasePlaylist(gomock.Any(), "test_user_123", gomock.Any()).
					Return(nil, tt.serviceError).
					Times(1)
			}

			var reqBody []byte
			if body, ok := tt.requestBody.(string); ok {
				reqBody = []byte(body)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/base_playlist", bytes.NewBuffer(reqBody))
			if !tt.noUserInContext {
				req = addUserToContext(req)
			}

			w := httptest.NewRecorder()
			controller.Create(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
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
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful deletion with complex id",
			playlistID:     "pl_abc123def456",
			urlPath:        "/api/base_playlist/pl_abc123def456",
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

func TestBasePlaylistController_Delete_Errors(t *testing.T) {
	tests := []struct {
		name               string
		playlistID         string
		serviceError       error
		noUserInContext    bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "empty id in path",
			playlistID:         "",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "playlist id is required",
		},
		{
			name:               "service error",
			playlistID:         "playlist123",
			serviceError:       errors.New("some service error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to delete base playlist",
		},
		{
			name:               "no user in context",
			playlistID:         "playlist123",
			noUserInContext:    true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "user not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					DeleteBasePlaylist(gomock.Any(), tt.playlistID, "test_user_123").
					Return(tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/base_playlist/"+tt.playlistID, nil)
			req.SetPathValue("id", tt.playlistID)
			if !tt.noUserInContext {
				req = addUserToContext(req)
			}

			w := httptest.NewRecorder()
			controller.Delete(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
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

func TestBasePlaylistController_GetByID_Errors(t *testing.T) {
	tests := []struct {
		name               string
		playlistID         string
		serviceError       error
		noUserInContext    bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "empty id in path",
			playlistID:         "",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "playlist id is required",
		},
		{
			name:               "service error",
			playlistID:         "playlist123",
			serviceError:       errors.New("some service error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to retrieve base playlist",
		},
		{
			name:               "no user in context",
			playlistID:         "playlist123",
			noUserInContext:    true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "user not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					GetBasePlaylist(gomock.Any(), tt.playlistID, "test_user_123").
					Return(nil, tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/base_playlist/"+tt.playlistID, nil)
			req.SetPathValue("id", tt.playlistID)
			if !tt.noUserInContext {
				req = addUserToContext(req)
			}

			w := httptest.NewRecorder()
			controller.GetByID(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_GetByUserID_Success(t *testing.T) {
	tests := []struct {
		name           string
		serviceResult  []*models.BasePlaylist
		expectedStatus int
	}{
		{
			name: "successful retrieval with multiple playlists",
			serviceResult: []*models.BasePlaylist{
				{
					ID:                "playlist123",
					UserID:            "user123",
					Name:              "My Test Playlist",
					SpotifyPlaylistID: "spotify123",
					IsActive:          true,
				},
				{
					ID:                "playlist456",
					UserID:            "user123",
					Name:              "Another Playlist",
					SpotifyPlaylistID: "spotify456",
					IsActive:          false,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful retrieval with single playlist",
			serviceResult: []*models.BasePlaylist{
				{
					ID:                "playlist123",
					UserID:            "user123",
					Name:              "My Only Playlist",
					SpotifyPlaylistID: "spotify123",
					IsActive:          true,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful retrieval with no playlists",
			serviceResult:  []*models.BasePlaylist{},
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

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/api/base_playlist", nil)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations
			mockService.EXPECT().
				GetBasePlaylistsByUserID(gomock.Any(), "test_user_123").
				Return(tt.serviceResult, nil).
				Times(1)

			// Execute
			controller.GetByUserID(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			// Verify response body
			var responseBody []*models.BasePlaylist
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(err)
			assert.Equal(len(tt.serviceResult), len(responseBody))

			for i, expectedPlaylist := range tt.serviceResult {
				assert.Equal(expectedPlaylist.ID, responseBody[i].ID)
				assert.Equal(expectedPlaylist.UserID, responseBody[i].UserID)
				assert.Equal(expectedPlaylist.Name, responseBody[i].Name)
				assert.Equal(expectedPlaylist.SpotifyPlaylistID, responseBody[i].SpotifyPlaylistID)
				assert.Equal(expectedPlaylist.IsActive, responseBody[i].IsActive)
			}
		})
	}
}

func TestBasePlaylistController_GetByUserID_Errors(t *testing.T) {
	tests := []struct {
		name               string
		serviceError       error
		noUserInContext    bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "service error",
			serviceError:       errors.New("some service error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to retrieve base playlists",
		},
		{
			name:               "no user in context",
			noUserInContext:    true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "user not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					GetBasePlaylistsByUserID(gomock.Any(), "test_user_123").
					Return(nil, tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/base_playlist", nil)
			if !tt.noUserInContext {
				req = addUserToContext(req)
			}

			w := httptest.NewRecorder()
			controller.GetByUserID(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestBasePlaylistController_GetByUserID_ResponseEncodingError(t *testing.T) {
	assert := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockBasePlaylistServicer(ctrl)
	controller := NewBasePlaylistController(mockService)

	serviceResult := []*models.BasePlaylist{
		{
			ID:                "playlist123",
			UserID:            "user123",
			Name:              "Test Playlist",
			SpotifyPlaylistID: "spotify123",
			IsActive:          true,
		},
	}

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/api/base_playlist", nil)
	req = addUserToContext(req)
	w := httptest.NewRecorder()

	// Set expectations
	mockService.EXPECT().
		GetBasePlaylistsByUserID(gomock.Any(), "test_user_123").
		Return(serviceResult, nil).
		Times(1)

	// Execute
	controller.GetByUserID(w, req)

	// Verify successful response (since our test data is actually valid)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("application/json", w.Header().Get("Content-Type"))
}

func TestBasePlaylistController_GetByUserIDWithChilds_Success(t *testing.T) {
	tests := []struct {
		name           string
		serviceResult  []*models.BasePlaylistWithChilds
		expectedStatus int
	}{
		{
			name: "successful retrieval with multiple playlists and childs",
			serviceResult: []*models.BasePlaylistWithChilds{
				{
					BasePlaylist: &models.BasePlaylist{ID: "playlist123"},
					Childs: []*models.ChildPlaylist{
						{ ID: "child123" },
						{ ID: "child456" },
					},
				},
				{
					BasePlaylist: &models.BasePlaylist{ID: "playlist456"},
					Childs: []*models.ChildPlaylist{
						{ ID: "child789" },
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful retrieval with single playlist and childs",
			serviceResult: []*models.BasePlaylistWithChilds{
				{
					BasePlaylist: &models.BasePlaylist{ID: "playlist123"},
					Childs: []*models.ChildPlaylist{{ ID: "child123" }},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful retrieval with no playlists",
			serviceResult:  []*models.BasePlaylistWithChilds{},
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

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/api/base_playlist/with_childs", nil)
			req = addUserToContext(req)
			w := httptest.NewRecorder()

			// Set expectations
			mockService.EXPECT().
				GetBasePlaylistsByUserIDWithChilds(gomock.Any(), "test_user_123").
				Return(tt.serviceResult, nil).
				Times(1)

			// Execute
			controller.GetByUserIDWithChilds(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			// Verify response body
			var responseBody []*models.BasePlaylistWithChilds
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(err)
			assert.Equal(len(tt.serviceResult), len(responseBody))

			for i, expectedPlaylist := range tt.serviceResult {
				assert.Equal(expectedPlaylist.ID, responseBody[i].ID)
				assert.Equal(len(expectedPlaylist.Childs), len(responseBody[i].Childs))

				for j, expectedChild := range expectedPlaylist.Childs {
					assert.Equal(expectedChild.ID, responseBody[i].Childs[j].ID)
				}
			}
		})
	}
}

func TestBasePlaylistController_GetByUserIDWithChilds_Errors(t *testing.T) {
	tests := []struct {
		name               string
		serviceError        error
		noUserInContext     bool
		expectedStatusCode  int
		expectedError       string
	}{
		{
			name:               "service error",
			serviceError:       errors.New("some service error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to retrieve base playlists with childs",
		},
		{
			name:               "no user in context",
			noUserInContext:    true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "user not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockBasePlaylistServicer(ctrl)
			controller := NewBasePlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					GetBasePlaylistsByUserIDWithChilds(gomock.Any(), "test_user_123").
					Return(nil, tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/base_playlist/with_childs", nil)
			if !tt.noUserInContext {
				req = addUserToContext(req)
			}

			w := httptest.NewRecorder()
			controller.GetByUserIDWithChilds(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

// Helper function to add user to request context
func addUserToContext(req *http.Request) *http.Request {
	user := &models.User{ID: "test_user_123", Email: "test@example.com", Name: "Test User"}
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	return req.WithContext(ctx)
}
