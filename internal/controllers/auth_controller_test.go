package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/ngomez18/playlist-router/internal/middleware"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func createTestConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			FrontendURL: "http://localhost:3000",
		},
	}
}

func TestAuthController_SpotifyLogin(t *testing.T) {
	tests := []struct {
		name               string
		expectedAuthURL    string
		expectedStatusCode int
	}{
		{
			name:               "successful login redirect",
			expectedAuthURL:    "https://accounts.spotify.com/authorize?client_id=test&state=somestate",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthService := mocks.NewMockAuthServicer(ctrl)
			cfg := createTestConfig()
			controller := NewAuthController(mockAuthService, cfg)

			// Setup mock expectations - we can't predict the exact state, so use Any()
			mockAuthService.EXPECT().
				GenerateSpotifyAuthURL(gomock.Any()).
				Return(tt.expectedAuthURL).
				Times(1)

			// Create request
			req := httptest.NewRequest("GET", "/auth/spotify/login", nil)
			w := httptest.NewRecorder()

			// Execute
			controller.SpotifyLogin(w, req)

			// Assert
			assert.Equal(tt.expectedStatusCode, w.Code)

			// Check redirect location
			location := w.Header().Get("Location")
			assert.Equal(tt.expectedAuthURL, location)
		})
	}
}

func TestAuthController_SpotifyCallback(t *testing.T) {
	tests := []struct {
		name                 string
		queryParams          map[string]string
		mockAuthResult       *services.AuthResult
		mockError            error
		expectedStatusCode   int
		expectedRedirectURL  string
		expectRedirect       bool
	}{
		{
			name: "successful callback",
			queryParams: map[string]string{
				"code":  "auth_code_123",
				"state": "state_123",
			},
			mockAuthResult: &services.AuthResult{
				User: &models.AuthUser{
					ID:        "user_123",
					Email:     "test@example.com",
					Name:      "Test User",
					SpotifyID: "spotify_user_123",
				},
				Token:        "pb_token_123",
				RefreshToken: "",
			},
			mockError:           nil,
			expectedStatusCode:  http.StatusTemporaryRedirect,
			expectedRedirectURL: "http://localhost:3000/?token=pb_token_123",
			expectRedirect:      true,
		},
		{
			name: "missing authorization code",
			queryParams: map[string]string{
				"state": "state_123",
			},
			mockAuthResult:     nil,
			mockError:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectRedirect:     false,
		},
		{
			name: "auth service error",
			queryParams: map[string]string{
				"code":  "invalid_code",
				"state": "state_123",
			},
			mockAuthResult:     nil,
			mockError:          errors.New("spotify API error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectRedirect:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthService := mocks.NewMockAuthServicer(ctrl)
			cfg := createTestConfig()
			controller := NewAuthController(mockAuthService, cfg)

			// Setup mock expectations (only if we have a code parameter)
			if code := tt.queryParams["code"]; code != "" {
				state := tt.queryParams["state"]
				mockAuthService.EXPECT().
					HandleSpotifyCallback(gomock.Any(), code, state).
					Return(tt.mockAuthResult, tt.mockError).
					Times(1)
			}

			// Create request with query parameters
			u := &url.URL{Path: "/auth/spotify/callback"}
			q := u.Query()
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			u.RawQuery = q.Encode()

			req := httptest.NewRequest("GET", u.String(), nil)
			w := httptest.NewRecorder()

			// Execute
			controller.SpotifyCallback(w, req)

			// Assert
			assert.Equal(tt.expectedStatusCode, w.Code)

			if tt.expectRedirect {
				// Check redirect location
				location := w.Header().Get("Location")
				assert.Equal(tt.expectedRedirectURL, location)
			}
		})
	}
}

func TestAuthController_SpotifyCallback_RedirectWithToken(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthServicer(ctrl)
	cfg := createTestConfig()
	controller := NewAuthController(mockAuthService, cfg)

	// Create a mock auth result
	mockAuthResult := &services.AuthResult{
		User: &models.AuthUser{
			ID:        "user_123",
			Email:     "test@example.com",
			Name:      "Test User",
			SpotifyID: "spotify_user_123",
		},
		Token:        "pb_token_123",
		RefreshToken: "",
	}

	// Setup mock expectations
	mockAuthService.EXPECT().
		HandleSpotifyCallback(gomock.Any(), "auth_code_123", "state_123").
		Return(mockAuthResult, nil).
		Times(1)

	// Create request
	req := httptest.NewRequest("GET", "/auth/spotify/callback?code=auth_code_123&state=state_123", nil)
	w := httptest.NewRecorder()

	// Execute
	controller.SpotifyCallback(w, req)

	// Assert redirect
	assert.Equal(http.StatusTemporaryRedirect, w.Code)
	expectedURL := "http://localhost:3000/?token=pb_token_123"
	assert.Equal(expectedURL, w.Header().Get("Location"))
}

func TestGenerateState(t *testing.T) {
	assert := require.New(t)

	// Test that generateState returns a non-empty hex string
	state1 := generateState()
	state2 := generateState()

	// Should be non-empty
	assert.NotEmpty(state1)
	assert.NotEmpty(state2)

	// Should be different on each call (extremely high probability)
	assert.NotEqual(state1, state2)

	// Should be 32 characters (16 bytes * 2 for hex encoding)
	assert.Equal(32, len(state1))
	assert.Equal(32, len(state2))

	// Should be valid hex
	for _, char := range state1 {
		assert.True(
			(char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
			"State should contain only hex characters, got: %c", char)
	}
}

func TestNewAuthController(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthServicer(ctrl)
	cfg := createTestConfig()
	controller := NewAuthController(mockAuthService, cfg)

	assert.NotNil(controller)
	assert.Equal(mockAuthService, controller.authService)
	assert.Equal(cfg, controller.config)
}


// Test that the controller properly handles context in requests
func TestAuthController_SpotifyCallback_ContextPropagation(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthServicer(ctrl)
	cfg := createTestConfig()
	controller := NewAuthController(mockAuthService, cfg)

	// Create a request with a custom context value
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("test_key"), "test_value")
	req := httptest.NewRequest("GET", "/auth/spotify/callback?code=test_code&state=test_state", nil)
	req = req.WithContext(ctx)

	mockAuthResult := &services.AuthResult{
		User: &models.AuthUser{
			ID:        "user_123",
			Email:     "test@example.com",
			Name:      "Test User",
			SpotifyID: "spotify_user_123",
		},
		Token:        "pb_token_123",
		RefreshToken: "",
	}

	// Setup mock expectations with context verification
	mockAuthService.EXPECT().
		HandleSpotifyCallback(gomock.Any(), "test_code", "test_state").
		Do(func(ctx context.Context, code, state string) {
			// Verify our custom context value is present
			assert.Equal("test_value", ctx.Value(ctxKey("test_key")))
		}).
		Return(mockAuthResult, nil).
		Times(1)

	w := httptest.NewRecorder()

	// Execute
	controller.SpotifyCallback(w, req)

	// Assert redirect
	assert.Equal(http.StatusTemporaryRedirect, w.Code)
	expectedURL := "http://localhost:3000/?token=pb_token_123"
	assert.Equal(expectedURL, w.Header().Get("Location"))
}

func TestAuthController_ValidateToken_Success(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		expectedStatus int
	}{
		{
			name: "successful token validation with complete user",
			user: &models.User{
				ID:    "user123",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful token validation with minimal user",
			user: &models.User{
				ID:    "user456",
				Email: "minimal@example.com",
				Name:  "Minimal User",
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

			mockAuthService := mocks.NewMockAuthServicer(ctrl)
			cfg := createTestConfig()
			controller := NewAuthController(mockAuthService, cfg)

			// Create request with user context
			req := httptest.NewRequest(http.MethodGet, "/auth/validate", nil)
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, tt.user)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			// No service expectations needed since this method does not call the service

			// Execute
			controller.ValidateToken(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Equal("application/json", w.Header().Get("Content-Type"))

			// Verify response body contains user data
			var responseUser models.User
			err := json.Unmarshal(w.Body.Bytes(), &responseUser)
			assert.NoError(err)
			assert.Equal(tt.user.ID, responseUser.ID)
			assert.Equal(tt.user.Email, responseUser.Email)
			assert.Equal(tt.user.Name, responseUser.Name)
		})
	}
}

func TestAuthController_ValidateToken_Unauthorized(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "user not found in context",
			setupContext: func(req *http.Request) *http.Request {
				// Return request without user context
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not found in context",
		},
		{
			name: "invalid user context type",
			setupContext: func(req *http.Request) *http.Request {
				// Add invalid type to context
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, "invalid_user_type")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthService := mocks.NewMockAuthServicer(ctrl)
			cfg := createTestConfig()
			controller := NewAuthController(mockAuthService, cfg)

			// Create request with specific context setup
			req := httptest.NewRequest(http.MethodGet, "/auth/validate", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			// No service expectations needed since auth should fail before service call

			// Execute
			controller.ValidateToken(w, req)

			// Verify response
			assert.Equal(tt.expectedStatus, w.Code)
			assert.Contains(w.Body.String(), tt.expectedBody)
		})
	}
}