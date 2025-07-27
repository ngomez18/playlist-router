package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/clients/mocks"
	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/require"
)

func TestNewSpotifyClient(t *testing.T) {
	assert := require.New(t)

	cfg := &config.AuthConfig{
		SpotifyClientID:     "test_client_id",
		SpotifyClientSecret: "test_client_secret",
		SpotifyRedirectURI:  "http://localhost:8080/callback",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client := NewSpotifyClient(cfg, logger)

	assert.NotNil(client)
	assert.NotNil(client.HttpClient)
	assert.Equal(cfg, client.config)
	assert.NotNil(client.logger)
}

func TestSpotifyClient_GenerateAuthURL(t *testing.T) {
	assert := require.New(t)

	clientID := "test_client_id"
	redirectURI := "http://localhost:8080/callback"
	cfg := &config.AuthConfig{
		SpotifyClientID:    clientID,
		SpotifyRedirectURI: redirectURI,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client := NewSpotifyClient(cfg, logger)

	state := "test_state"
	authURL := client.GenerateAuthURL(state)

	// Parse the URL to validate components
	parsedURL, err := url.Parse(authURL)
	assert.NoError(err)

	assert.Equal("accounts.spotify.com", parsedURL.Host)
	assert.Equal("/authorize", parsedURL.Path)

	// Check query parameters
	params := parsedURL.Query()
	assert.Equal(state, params.Get("state"))
	assert.Equal(clientID, params.Get("client_id"))
	assert.Equal(redirectURI, params.Get("redirect_uri"))
	assert.Equal("code", params.Get("response_type"))
	assert.Equal("user-read-email playlist-read-private playlist-modify-public playlist-modify-private", params.Get("scope"))
}

func TestSpotifyClient_ExchangeCodeForTokens(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   *models.SpotifyTokenResponse
		responseError  error
		expectError    bool
		expectedTokens *models.SpotifyTokenResponse
	}{
		{
			name:           "successful token exchange",
			responseStatus: http.StatusOK,
			responseBody: &models.SpotifyTokenResponse{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				Scope:        "user-read-email",
				ExpiresIn:    3600,
				RefreshToken: "refresh_token_123",
			},
			expectError: false,
			expectedTokens: &models.SpotifyTokenResponse{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				Scope:        "user-read-email",
				ExpiresIn:    3600,
				RefreshToken: "refresh_token_123",
			},
		},
		{
			name:           "spotify API error",
			responseStatus: http.StatusBadRequest,
			responseBody:   nil,
			expectError:    true,
		},
		{
			name:          "http client error",
			responseError: errors.New("http client error"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{
				SpotifyClientID:     "test_client_id",
				SpotifyClientSecret: "test_client_secret",
				SpotifyRedirectURI:  "http://localhost:8080/callback",
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// Create client and overwrite HTTP client with mock
			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Setup mock expectations
			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} else {
				// Create response body
				var responseBody io.ReadCloser
				if tt.responseBody != nil {
					bodyBytes, _ := json.Marshal(tt.responseBody)
					responseBody = io.NopCloser(bytes.NewReader(bodyBytes))
				} else {
					responseBody = io.NopCloser(strings.NewReader(`{"error":"invalid_grant"}`))
				}

				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Validate request
						assert.Equal("POST", req.Method)
						assert.Equal("https://accounts.spotify.com/api/token", req.URL.String())
						assert.Equal("application/x-www-form-urlencoded", req.Header.Get("Content-Type"))

						// Check basic auth
						username, password, ok := req.BasicAuth()
						assert.True(ok)
						assert.Equal("test_client_id", username)
						assert.Equal("test_client_secret", password)

						// Check form data
						body, _ := io.ReadAll(req.Body)
						form, _ := url.ParseQuery(string(body))
						assert.Equal("authorization_code", form.Get("grant_type"))
						assert.Equal("test_code", form.Get("code"))
						assert.Equal("http://localhost:8080/callback", form.Get("redirect_uri"))

						return resp, nil
					}).
					Times(1)
			}

			ctx := context.Background()
			tokens, err := client.ExchangeCodeForTokens(ctx, "test_code")

			if tt.expectError {
				assert.Error(err)
				assert.Nil(tokens)
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedTokens, tokens)
			}
		})
	}
}

func TestSpotifyClient_GetUserProfile(t *testing.T) {
	tests := []struct {
		name            string
		responseStatus  int
		responseBody    *models.SpotifyUserProfile
		responseError   error
		expectError     bool
		expectedProfile *models.SpotifyUserProfile
		accessToken     string
	}{
		{
			name:           "successful profile fetch",
			responseStatus: http.StatusOK,
			responseBody: &models.SpotifyUserProfile{
				ID:    "user123",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expectError: false,
			expectedProfile: &models.SpotifyUserProfile{
				ID:    "user123",
				Email: "test@example.com",
				Name:  "Test User",
			},
			accessToken: "valid_token",
		},
		{
			name:           "unauthorized error",
			responseStatus: http.StatusUnauthorized,
			responseBody:   nil,
			expectError:    true,
			accessToken:    "invalid_token",
		},
		{
			name:          "http client error",
			responseError: errors.New("http client error"),
			expectError:   true,
			accessToken:   "valid_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			cfg := &config.AuthConfig{
				SpotifyClientID:     "test_client_id",
				SpotifyClientSecret: "test_client_secret",
				SpotifyRedirectURI:  "http://localhost:8080/callback",
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// Create client and overwrite HTTP client with mock
			client := NewSpotifyClient(cfg, logger)
			client.HttpClient = mockHTTPClient

			// Setup mock expectations
			if tt.responseError != nil {
				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					Return(nil, tt.responseError).
					Times(1)
			} else {
				// Create response body
				var responseBody io.ReadCloser
				if tt.responseBody != nil {
					bodyBytes, _ := json.Marshal(tt.responseBody)
					responseBody = io.NopCloser(bytes.NewReader(bodyBytes))
				} else {
					responseBody = io.NopCloser(strings.NewReader(`{"error":"invalid_token"}`))
				}

				resp := &http.Response{
					StatusCode: tt.responseStatus,
					Body:       responseBody,
				}

				mockHTTPClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						// Validate request
						assert.Equal("GET", req.Method)
						assert.Equal("https://api.spotify.com/v1/me", req.URL.String())
						assert.Equal("Bearer "+tt.accessToken, req.Header.Get("Authorization"))

						return resp, nil
					}).
					Times(1)
			}

			ctx := context.Background()
			profile, err := client.GetUserProfile(ctx, tt.accessToken)

			if tt.expectError {
				assert.Error(err)
				assert.Nil(profile)
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedProfile, profile)
			}
		})
	}
}
