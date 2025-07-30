package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	spotifymocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	servicemocks "github.com/ngomez18/playlist-router/internal/services/mocks"
)

func TestNewSpotifyAuthMiddleware(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyService := servicemocks.NewMockSpotifyIntegrationServicer(ctrl)
	mockSpotifyClient := spotifymocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	middleware := NewSpotifyAuthMiddleware(mockSpotifyService, mockSpotifyClient, logger)

	assert.NotNil(middleware)
	assert.Equal(mockSpotifyService, middleware.spotifyIntegrationService)
	assert.Equal(mockSpotifyClient, middleware.spotifyClient)
	assert.NotNil(middleware.logger)
}

func TestSpotifyAuthMiddleware_RequireSpotifyAuth_Success(t *testing.T) {
	tests := []struct {
		name           string
		tokenExpiry    time.Time
		shouldRefresh  bool
		refreshSuccess bool
	}{
		{
			name:           "valid token, no refresh needed",
			tokenExpiry:    time.Now().Add(1 * time.Hour),
			shouldRefresh:  false,
			refreshSuccess: false,
		},
		{
			name:           "token expires soon, successful refresh",
			tokenExpiry:    time.Now().Add(5 * time.Minute),
			shouldRefresh:  true,
			refreshSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyService := servicemocks.NewMockSpotifyIntegrationServicer(ctrl)
			mockSpotifyClient := spotifymocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			middleware := NewSpotifyAuthMiddleware(mockSpotifyService, mockSpotifyClient, logger)

			// Create test user and integration
			user := &models.User{ID: "user123"}
			integration := &models.SpotifyIntegration{
				ID:           "integration123",
				UserID:       "user123",
				SpotifyID:    "spotify123",
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
				ExpiresAt:    tt.tokenExpiry,
			}

			// Mock service calls
			mockSpotifyService.EXPECT().
				GetIntegrationByUserID(gomock.Any(), "user123").
				Return(integration, nil).
				Times(1)

			if tt.shouldRefresh {
				// Mock token refresh
				refreshResponse := &spotifyclient.SpotifyTokenResponse{
					AccessToken:  "new_access_token_456",
					RefreshToken: "new_refresh_token_456",
					ExpiresIn:    3600,
				}

				mockSpotifyClient.EXPECT().
					RefreshTokens(gomock.Any(), "refresh_token_123").
					Return(refreshResponse, nil).
					Times(1)

				mockSpotifyService.EXPECT().
					UpdateTokens(gomock.Any(), "integration123", gomock.Any()).
					DoAndReturn(func(ctx context.Context, integrationID string, tokens *models.SpotifyIntegrationTokenRefresh) error {
						assert.Equal("new_access_token_456", tokens.AccessToken)
						assert.Equal("new_refresh_token_456", tokens.RefreshToken)
						assert.Equal(3600, tokens.ExpiresIn)
						return nil
					}).
					Times(1)
			}

			// Create request with user in context
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := requestcontext.ContextWithUser(req.Context(), user)
			req = req.WithContext(ctx)

			// Create handler that validates the context
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				spotifyIntegration, ok := requestcontext.GetSpotifyAuthFromContext(r.Context())
				assert.True(ok)
				assert.NotNil(spotifyIntegration)
				assert.Equal("user123", spotifyIntegration.UserID)

				if tt.shouldRefresh {
					// Token should be refreshed
					assert.Equal("new_access_token_456", spotifyIntegration.AccessToken)
					assert.Equal("new_refresh_token_456", spotifyIntegration.RefreshToken)
					assert.True(spotifyIntegration.ExpiresAt.After(time.Now().Add(59 * time.Minute)))
				} else {
					// Token should be unchanged
					assert.Equal("access_token_123", spotifyIntegration.AccessToken)
					assert.Equal("refresh_token_123", spotifyIntegration.RefreshToken)
				}

				w.WriteHeader(http.StatusOK)
			})

			// Execute middleware
			w := httptest.NewRecorder()
			middleware.RequireSpotifyAuth(testHandler).ServeHTTP(w, req)

			assert.Equal(http.StatusOK, w.Code)
			assert.True(handlerCalled)
		})
	}
}

func TestSpotifyAuthMiddleware_RequireSpotifyAuth_TokenRefreshNoNewRefreshToken(t *testing.T) {
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyIntegrationService := servicemocks.NewMockSpotifyIntegrationServicer(ctrl)
	mockSpotifyClient := spotifymocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	middleware := NewSpotifyAuthMiddleware(mockSpotifyIntegrationService, mockSpotifyClient, logger)

	// Create test user and integration
	user := &models.User{ID: "user123"}
	integration := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       "user123",
		SpotifyID:    "spotify123",
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Expires soon
	}

	// Mock service calls
	mockSpotifyIntegrationService.EXPECT().
		GetIntegrationByUserID(gomock.Any(), "user123").
		Return(integration, nil).
		Times(1)

	// Mock token refresh without new refresh token
	refreshResponse := &spotifyclient.SpotifyTokenResponse{
		AccessToken:  "new_access_token_456",
		RefreshToken: "", // No new refresh token
		ExpiresIn:    3600,
	}

	mockSpotifyClient.EXPECT().
		RefreshTokens(gomock.Any(), "refresh_token_123").
		Return(refreshResponse, nil).
		Times(1)

	mockSpotifyIntegrationService.EXPECT().
		UpdateTokens(gomock.Any(), "integration123", gomock.Any()).
		DoAndReturn(func(ctx context.Context, integrationID string, tokens *models.SpotifyIntegrationTokenRefresh) error {
			assert.Equal("new_access_token_456", tokens.AccessToken)
			assert.Equal("refresh_token_123", tokens.RefreshToken) // Should keep original
			assert.Equal(3600, tokens.ExpiresIn)
			return nil
		}).
		Times(1)

	// Create request with user in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := requestcontext.ContextWithUser(req.Context(), user)
	req = req.WithContext(ctx)

	// Create handler that validates the context
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		spotifyIntegration, ok := requestcontext.GetSpotifyAuthFromContext(r.Context())
		assert.True(ok)
		assert.NotNil(spotifyIntegration)
		assert.Equal("new_access_token_456", spotifyIntegration.AccessToken)
		assert.Equal("refresh_token_123", spotifyIntegration.RefreshToken) // Should keep original
		w.WriteHeader(http.StatusOK)
	})

	// Execute middleware
	w := httptest.NewRecorder()
	middleware.RequireSpotifyAuth(testHandler).ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.True(handlerCalled)
}

func TestSpotifyAuthMiddleware_RequireSpotifyAuth_Errors(t *testing.T) {
	tests := []struct {
		name               string
		userInContext      bool
		integrationError   error
		tokenRefreshError  error
		dbUpdateError      error
		expectedStatusCode int
	}{
		{
			name:               "no user in context",
			userInContext:      false,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "spotify integration not found",
			userInContext:      true,
			integrationError:   assert.AnError,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "token refresh fails",
			userInContext:      true,
			tokenRefreshError:  assert.AnError,
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyIntegrationService := servicemocks.NewMockSpotifyIntegrationServicer(ctrl)
			mockSpotifyClient := spotifymocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			middleware := NewSpotifyAuthMiddleware(mockSpotifyIntegrationService, mockSpotifyClient, logger)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)

			if tt.userInContext {
				user := &models.User{ID: "user123"}
				ctx := requestcontext.ContextWithUser(req.Context(), user)
				req = req.WithContext(ctx)

				if tt.integrationError != nil {
					mockSpotifyIntegrationService.EXPECT().
						GetIntegrationByUserID(gomock.Any(), "user123").
						Return(nil, tt.integrationError).
						Times(1)
				} else {
					integration := &models.SpotifyIntegration{
						ID:           "integration123",
						UserID:       "user123",
						SpotifyID:    "spotify123",
						AccessToken:  "access_token_123",
						RefreshToken: "refresh_token_123",
						ExpiresAt:    time.Now().Add(5 * time.Minute), // Needs refresh
					}

					mockSpotifyIntegrationService.EXPECT().
						GetIntegrationByUserID(gomock.Any(), "user123").
						Return(integration, nil).
						Times(1)

					if tt.tokenRefreshError != nil {
						mockSpotifyClient.EXPECT().
							RefreshTokens(gomock.Any(), "refresh_token_123").
							Return(nil, tt.tokenRefreshError).
							Times(1)
					} else if tt.dbUpdateError != nil {
						refreshResponse := &spotifyclient.SpotifyTokenResponse{
							AccessToken:  "new_access_token_456",
							RefreshToken: "new_refresh_token_456",
							ExpiresIn:    3600,
						}

						mockSpotifyClient.EXPECT().
							RefreshTokens(gomock.Any(), "refresh_token_123").
							Return(refreshResponse, nil).
							Times(1)

						mockSpotifyIntegrationService.EXPECT().
							UpdateTokens(gomock.Any(), "integration123", gomock.Any()).
							Return(tt.dbUpdateError).
							Times(1)
					}
				}
			}

			// Create test handler
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Execute middleware
			w := httptest.NewRecorder()
			middleware.RequireSpotifyAuth(testHandler).ServeHTTP(w, req)

			assert.Equal(tt.expectedStatusCode, w.Code)
			assert.False(handlerCalled)
		})
	}
}
