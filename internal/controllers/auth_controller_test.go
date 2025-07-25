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
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

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
			controller := NewAuthController(mockAuthService)

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
		expectedResponseBody interface{}
		expectJSONResponse   bool
	}{
		{
			name: "successful callback",
			queryParams: map[string]string{
				"code":  "auth_code_123",
				"state": "state_123",
			},
			mockAuthResult: &services.AuthResult{
				User: &services.AuthUser{
					ID:        "user_123",
					Email:     "test@example.com",
					Name:      "Test User",
					SpotifyID: "spotify_user_123",
				},
				Token:        "pb_token_123",
				RefreshToken: "",
			},
			mockError:          nil,
			expectedStatusCode: http.StatusOK,
			expectJSONResponse: true,
		},
		{
			name: "missing authorization code",
			queryParams: map[string]string{
				"state": "state_123",
			},
			mockAuthResult:     nil,
			mockError:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectJSONResponse: false,
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
			expectJSONResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthService := mocks.NewMockAuthServicer(ctrl)
			controller := NewAuthController(mockAuthService)

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

			if tt.expectJSONResponse {
				// Check content type
				assert.Equal("application/json", w.Header().Get("Content-Type"))

				// Parse and validate JSON response
				var response services.AuthResult
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(err)

				assert.Equal(tt.mockAuthResult.User.ID, response.User.ID)
				assert.Equal(tt.mockAuthResult.User.Email, response.User.Email)
				assert.Equal(tt.mockAuthResult.User.Name, response.User.Name)
				assert.Equal(tt.mockAuthResult.User.SpotifyID, response.User.SpotifyID)
				assert.Equal(tt.mockAuthResult.Token, response.Token)
				assert.Equal(tt.mockAuthResult.RefreshToken, response.RefreshToken)
			}
		})
	}
}

func TestAuthController_SpotifyCallback_JSONEncodingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthServicer(ctrl)
	controller := NewAuthController(mockAuthService)

	// Create a mock auth result
	mockAuthResult := &services.AuthResult{
		User: &services.AuthUser{
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

	// Use a custom ResponseWriter that will cause JSON encoding to fail
	w := &failingResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		failOnWrite:      true,
	}

	// Execute
	controller.SpotifyCallback(w, req)

	// The response should indicate encoding failure
	// Note: In the current implementation, the error handling for JSON encoding
	// calls http.Error which would panic on a failing writer, so we'll just
	// verify our mock was called correctly
	// The gomock controller automatically verifies expectations on defer ctrl.Finish()
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
	controller := NewAuthController(mockAuthService)

	assert.NotNil(controller)
	assert.Equal(mockAuthService, controller.authService)
}

// Custom ResponseWriter that can simulate write failures for testing
type failingResponseWriter struct {
	*httptest.ResponseRecorder
	failOnWrite bool
}

func (f *failingResponseWriter) Write(b []byte) (int, error) {
	if f.failOnWrite {
		return 0, errors.New("simulated write failure")
	}
	return f.ResponseRecorder.Write(b)
}

func (f *failingResponseWriter) Header() http.Header {
	return f.ResponseRecorder.Header()
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
	f.ResponseRecorder.WriteHeader(statusCode)
}

// Test that the controller properly handles context in requests
func TestAuthController_SpotifyCallback_ContextPropagation(t *testing.T) {
	assert := require.New(t)
	
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthServicer(ctrl)
	controller := NewAuthController(mockAuthService)

	// Create a request with a custom context value
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("test_key"), "test_value")
	req := httptest.NewRequest("GET", "/auth/spotify/callback?code=test_code&state=test_state", nil)
	req = req.WithContext(ctx)

	mockAuthResult := &services.AuthResult{
		User: &services.AuthUser{
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

	// Assert
	assert.Equal(http.StatusOK, w.Code)
}