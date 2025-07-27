package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/models"
	serviceMocks "github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedUser  *models.User
		expectedFound bool
	}{
		{
			name: "user_found_in_context",
			setupContext: func() context.Context {
				user := &models.User{
					ID:       "user123",
					Email:    "test@example.com",
					Username: "testuser",
					Name:     "Test User",
				}
				return context.WithValue(context.Background(), UserContextKey, user)
			},
			expectedUser: &models.User{
				ID:       "user123",
				Email:    "test@example.com",
				Username: "testuser",
				Name:     "Test User",
			},
			expectedFound: true,
		},
		{
			name: "user_not_found_in_context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedUser:  nil,
			expectedFound: false,
		},
		{
			name: "wrong_type_in_context",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), UserContextKey, "not a user")
			},
			expectedUser:  nil,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := tt.setupContext()
			user, found := GetUserFromContext(ctx)

			assert.Equal(tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(tt.expectedUser.ID, user.ID)
				assert.Equal(tt.expectedUser.Email, user.Email)
				assert.Equal(tt.expectedUser.Username, user.Username)
				assert.Equal(tt.expectedUser.Name, user.Name)
			} else {
				assert.Nil(user)
			}
		})
	}
}

func TestAuthMiddleware_RequireAuth_Errors(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(*serviceMocks.MockUserServicer)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing_auth_header",
			authHeader:     "",
			setupMock:      func(mock *serviceMocks.MockUserServicer) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "authorization header is required",
		},
		{
			name:           "invalid_format_no_bearer",
			authHeader:     "Invalid token123",
			setupMock:      func(mock *serviceMocks.MockUserServicer) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid authorization header format",
		},
		{
			name:           "empty_token",
			authHeader:     "Bearer ",
			setupMock:      func(mock *serviceMocks.MockUserServicer) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "token is required",
		},
		{
			name:           "bearer_only",
			authHeader:     "Bearer",
			setupMock:      func(mock *serviceMocks.MockUserServicer) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid authorization header format",
		},
		{
			name:       "invalid_token",
			authHeader: "Bearer invalid_token",
			setupMock: func(mock *serviceMocks.MockUserServicer) {
				mock.EXPECT().
					ValidateAuthToken(gomock.Any(), "invalid_token").
					Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserService := serviceMocks.NewMockUserServicer(ctrl)
			tt.setupMock(mockUserService)
			middleware := NewAuthMiddleware(mockUserService)

			handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("success"))
				assert.NoError(err)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(tt.expectedStatus, recorder.Code)
			assert.Contains(recorder.Body.String(), tt.expectedError)
		})
	}
}

func TestAuthMiddleware_OptionalAuth_Success(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(*serviceMocks.MockUserServicer) *models.User
		expectUser     bool
	}{
		{
			name:       "no_auth_header",
			authHeader: "",
			setupMock: func(mock *serviceMocks.MockUserServicer) *models.User {
				return nil
			},
			expectUser: false,
		},
		{
			name:       "valid_token",
			authHeader: "Bearer valid_token",
			setupMock: func(mock *serviceMocks.MockUserServicer) *models.User {
				expectedUser := &models.User{
					ID:       "user123",
					Email:    "test@example.com",
					Username: "testuser",
					Name:     "Test User",
					Created:  time.Now(),
					Updated:  time.Now(),
				}
				mock.EXPECT().
					ValidateAuthToken(gomock.Any(), "valid_token").
					Return(expectedUser, nil)
				return expectedUser
			},
			expectUser: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserService := serviceMocks.NewMockUserServicer(ctrl)
			expectedUser := tt.setupMock(mockUserService)
			middleware := NewAuthMiddleware(mockUserService)
			handlerCalled := false

			handler := middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				user, found := GetUserFromContext(r.Context())
				
				if tt.expectUser {
					assert.True(found)
					assert.Equal(expectedUser.ID, user.ID)
				} else {
					assert.False(found)
					assert.Nil(user)
				}
				
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("success"))
				assert.NoError(err)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.True(handlerCalled)
			assert.Equal(http.StatusOK, recorder.Code)
			assert.Equal("success", recorder.Body.String())
		})
	}
}

func TestAuthMiddleware_ExtractUserID_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := &models.User{
		ID:       "user123",
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
	}

	ctx := context.WithValue(context.Background(), UserContextKey, user)
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)

	mockUserService := serviceMocks.NewMockUserServicer(ctrl)
	middleware := NewAuthMiddleware(mockUserService)
	userID, err := middleware.ExtractUserID(req)

	assert.NoError(err)
	assert.Equal("user123", userID)
}

func TestAuthMiddleware_ExtractUserID_NoUser(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := httptest.NewRequest("GET", "/test", nil)
	// No user in context

	mockUserService := serviceMocks.NewMockUserServicer(ctrl)
	middleware := NewAuthMiddleware(mockUserService)
	userID, err := middleware.ExtractUserID(req)

	assert.Error(err)
	assert.Empty(userID)
	assert.Contains(err.Error(), "User not found in context")
}

func TestAuthMiddleware_RequireAuth_ValidToken(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := serviceMocks.NewMockUserServicer(ctrl)
	middleware := NewAuthMiddleware(mockUserService)

	expectedUser := &models.User{
		ID:       "user123",
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	mockUserService.EXPECT().
		ValidateAuthToken(gomock.Any(), "valid_token").
		Return(expectedUser, nil)

	handlerCalled := false
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Check that user is in context
		user, found := GetUserFromContext(r.Context())
		assert.True(found)
		assert.Equal(expectedUser.ID, user.ID)
		assert.Equal(expectedUser.Email, user.Email)
		
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("success"))
		assert.NoError(err)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.True(handlerCalled)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("success", recorder.Body.String())
}



func TestAuthMiddleware_OptionalAuth_InvalidToken(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := serviceMocks.NewMockUserServicer(ctrl)
	middleware := NewAuthMiddleware(mockUserService)

	mockUserService.EXPECT().
		ValidateAuthToken(gomock.Any(), "invalid_token").
		Return(nil, errors.New("invalid token"))

	handlerCalled := false
	handler := middleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Check that no user is in context (failed auth continues without user)
		user, found := GetUserFromContext(r.Context())
		assert.False(found)
		assert.Nil(user)
		
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("success"))
		assert.NoError(err)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.True(handlerCalled)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("success", recorder.Body.String())
}