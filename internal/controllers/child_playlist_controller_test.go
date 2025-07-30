package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services/mocks"
)

func TestNewChildPlaylistController(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockChildPlaylistServicer(ctrl)
	controller := NewChildPlaylistController(mockService)

	assert.NotNil(controller)
	assert.Equal(mockService, controller.childPlaylistService)
	assert.NotNil(controller.validator)
}

func TestChildPlaylistController_Create_Success(t *testing.T) {
	tests := []struct {
		name               string
		basePlaylistID     string
		request            models.CreateChildPlaylistRequest
		serviceResult      *models.ChildPlaylist
		expectedStatusCode int
	}{
		{
			name:           "successful creation with valid request",
			basePlaylistID: "base123",
			request: models.CreateChildPlaylistRequest{
				Name:        "Test Child Playlist",
				Description: "Test description",
				FilterRules: &models.AudioFeatureFilters{
					Energy: &models.RangeFilter{Min: ptrFloat64(0.5), Max: ptrFloat64(1.0)},
				},
			},
			serviceResult: &models.ChildPlaylist{
				ID:                "child123",
				UserID:            "user123",
				BasePlaylistID:    "base123",
				Name:              "Test Child Playlist",
				Description:       "Test description",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:           "successful creation with minimal request",
			basePlaylistID: "base456",
			request: models.CreateChildPlaylistRequest{
				Name: "Minimal Child",
			},
			serviceResult: &models.ChildPlaylist{
				ID:                "child456",
				UserID:            "user123",
				BasePlaylistID:    "base456",
				Name:              "Minimal Child",
				SpotifyPlaylistID: "spotify456",
				IsActive:          true,
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			// Mock service expectation
			mockService.EXPECT().
				CreateChildPlaylist(gomock.Any(), "user123", tt.basePlaylistID, &tt.request).
				Return(tt.serviceResult, nil).
				Times(1)

			// Create request body
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/base_playlist/"+tt.basePlaylistID+"/child_playlist", bytes.NewReader(requestBody))
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("basePlaylistID", tt.basePlaylistID)

			w := httptest.NewRecorder()
			controller.Create(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			var response models.ChildPlaylist
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(err)
			assert.Equal(tt.serviceResult.ID, response.ID)
			assert.Equal(tt.serviceResult.Name, response.Name)
		})
	}
}

func TestChildPlaylistController_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name               string
		basePlaylistID     string
		request            models.CreateChildPlaylistRequest
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:           "empty name",
			basePlaylistID: "base123",
			request: models.CreateChildPlaylistRequest{
				Name: "",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "validation failed:",
		},
		{
			name:           "name too long",
			basePlaylistID: "base123",
			request: models.CreateChildPlaylistRequest{
				Name: "This is a very long name that exceeds the maximum allowed length for a child playlist name which should be 100 characters max",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "validation failed:",
		},
		{
			name:           "missing base playlist ID",
			basePlaylistID: "",
			request: models.CreateChildPlaylistRequest{
				Name: "Valid Name",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "base playlist ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			// Create request body
			requestBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/base_playlist/"+tt.basePlaylistID+"/child_playlist", bytes.NewReader(requestBody))
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("basePlaylistID", tt.basePlaylistID)

			w := httptest.NewRecorder()
			controller.Create(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestChildPlaylistController_Create_RequestParsingErrors(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "invalid JSON",
			requestBody:        `{"name": "test", invalid}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "invalid payload",
		},
		{
			name:               "empty body",
			requestBody:        "",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "invalid payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			req := httptest.NewRequest("POST", "/api/base_playlist/base123/child_playlist", bytes.NewReader([]byte(tt.requestBody)))
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("basePlaylistID", "base123")

			w := httptest.NewRecorder()
			controller.Create(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestChildPlaylistController_Create_ServiceErrors(t *testing.T) {
	tests := []struct {
		name               string
		serviceError       error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "base playlist not found",
			serviceError:       errors.New("base playlist not found"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to create child playlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			request := models.CreateChildPlaylistRequest{
				Name: "Test Child",
			}

			mockService.EXPECT().
				CreateChildPlaylist(gomock.Any(), "user123", "base123", &request).
				Return(nil, tt.serviceError).
				Times(1)

			requestBody, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/api/base_playlist/base123/child_playlist", bytes.NewReader(requestBody))
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("basePlaylistID", "base123")

			w := httptest.NewRecorder()
			controller.Create(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestChildPlaylistController_Create_AuthorizationErrors(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockChildPlaylistServicer(ctrl)
	controller := NewChildPlaylistController(mockService)

	request := models.CreateChildPlaylistRequest{
		Name: "Test Child",
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/base_playlist/base123/child_playlist", bytes.NewReader(requestBody))
	// No user in context
	req.SetPathValue("basePlaylistID", "base123")

	w := httptest.NewRecorder()
	controller.Create(w, req)

	assert.Equal(http.StatusUnauthorized, w.Code)
	assert.Contains(w.Body.String(), "user not found in context")
}

func TestChildPlaylistController_GetByID_Success(t *testing.T) {
	tests := []struct {
		name               string
		childPlaylistID    string
		serviceResult      *models.ChildPlaylist
		expectedStatusCode int
	}{
		{
			name:            "successful retrieval",
			childPlaylistID: "child123",
			serviceResult: &models.ChildPlaylist{
				ID:                "child123",
				UserID:            "user123",
				BasePlaylistID:    "base123",
				Name:              "Test Child Playlist",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			mockService.EXPECT().
				GetChildPlaylist(gomock.Any(), tt.childPlaylistID, "user123").
				Return(tt.serviceResult, nil).
				Times(1)

			req := httptest.NewRequest("GET", "/api/child_playlist/"+tt.childPlaylistID, nil)
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("id", tt.childPlaylistID)

			w := httptest.NewRecorder()
			controller.GetByID(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			var response models.ChildPlaylist
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(err)
			assert.Equal(tt.serviceResult.ID, response.ID)
		})
	}
}

func TestChildPlaylistController_GetByID_Errors(t *testing.T) {
	tests := []struct {
		name               string
		childPlaylistID    string
		serviceError       error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "child playlist not found",
			childPlaylistID:    "nonexistent",
			serviceError:       errors.New("child playlist not found"),
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "child playlist not found",
		},
		{
			name:               "empty ID",
			childPlaylistID:    "",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "child playlist ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					GetChildPlaylist(gomock.Any(), tt.childPlaylistID, "user123").
					Return(nil, tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest("GET", "/api/child_playlist/"+tt.childPlaylistID, nil)
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("id", tt.childPlaylistID)

			w := httptest.NewRecorder()
			controller.GetByID(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

func TestChildPlaylistController_GetByBasePlaylistID_Success(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockChildPlaylistServicer(ctrl)
	controller := NewChildPlaylistController(mockService)

	expectedPlaylists := []*models.ChildPlaylist{
		{
			ID:                "child1",
			UserID:            "user123",
			BasePlaylistID:    "base123",
			Name:              "Child 1",
			SpotifyPlaylistID: "spotify1",
		},
		{
			ID:                "child2",
			UserID:            "user123",
			BasePlaylistID:    "base123",
			Name:              "Child 2",
			SpotifyPlaylistID: "spotify2",
		},
	}

	mockService.EXPECT().
		GetChildPlaylistsByBasePlaylistID(gomock.Any(), "base123", "user123").
		Return(expectedPlaylists, nil).
		Times(1)

	req := httptest.NewRequest("GET", "/api/base_playlist/base123/child_playlist", nil)
	req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
	req.SetPathValue("basePlaylistID", "base123")

	w := httptest.NewRecorder()
	controller.GetByBasePlaylistID(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("application/json", w.Header().Get("Content-Type"))

	var response []*models.ChildPlaylist
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(err)
	assert.Len(response, 2)
	assert.Equal("child1", response[0].ID)
	assert.Equal("child2", response[1].ID)
}

func TestChildPlaylistController_Update_Success(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockChildPlaylistServicer(ctrl)
	controller := NewChildPlaylistController(mockService)

	newName := "Updated Name"
	newDescription := "Updated Description"
	request := models.UpdateChildPlaylistRequest{
		Name:        &newName,
		Description: &newDescription,
	}

	expectedResult := &models.ChildPlaylist{
		ID:                "child123",
		UserID:            "user123",
		BasePlaylistID:    "base123",
		Name:              "Updated Name",
		Description:       "Updated Description",
		SpotifyPlaylistID: "spotify123",
	}

	mockService.EXPECT().
		UpdateChildPlaylist(gomock.Any(), "child123", "user123", &request).
		Return(expectedResult, nil).
		Times(1)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("PUT", "/api/child_playlist/child123", bytes.NewReader(requestBody))
	req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
	req.SetPathValue("id", "child123")

	w := httptest.NewRecorder()
	controller.Update(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("application/json", w.Header().Get("Content-Type"))

	var response models.ChildPlaylist
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(err)
	assert.Equal("Updated Name", response.Name)
	assert.Equal("Updated Description", response.Description)
}

// TODO: Update error cases

func TestChildPlaylistController_Delete_Success(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockChildPlaylistServicer(ctrl)
	controller := NewChildPlaylistController(mockService)

	mockService.EXPECT().
		DeleteChildPlaylist(gomock.Any(), "child123", "user123").
		Return(nil).
		Times(1)

	req := httptest.NewRequest("DELETE", "/api/child_playlist/child123", nil)
	req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
	req.SetPathValue("id", "child123")

	w := httptest.NewRecorder()
	controller.Delete(w, req)

	assert.Equal(http.StatusNoContent, w.Code)
	assert.Empty(w.Body.String())
}

func TestChildPlaylistController_Delete_Errors(t *testing.T) {
	tests := []struct {
		name               string
		childPlaylistID    string
		serviceError       error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "child playlist not found",
			childPlaylistID:    "nonexistent",
			serviceError:       errors.New("child playlist not found"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "unable to delete child playlist",
		},
		{
			name:               "empty ID",
			childPlaylistID:    "",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "child playlist ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockChildPlaylistServicer(ctrl)
			controller := NewChildPlaylistController(mockService)

			if tt.serviceError != nil {
				mockService.EXPECT().
					DeleteChildPlaylist(gomock.Any(), tt.childPlaylistID, "user123").
					Return(tt.serviceError).
					Times(1)
			}

			req := httptest.NewRequest("DELETE", "/api/child_playlist/"+tt.childPlaylistID, nil)
			req = req.WithContext(requestcontext.ContextWithUser(req.Context(), &models.User{ID: "user123"}))
			req.SetPathValue("id", tt.childPlaylistID)

			w := httptest.NewRecorder()
			controller.Delete(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.Contains(w.Body.String(), tt.expectedError)
		})
	}
}

// ptrFloat64 returns a pointer to a float64 value
func ptrFloat64(f float64) *float64 {
	return &f
}
