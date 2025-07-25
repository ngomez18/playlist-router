# Authentication Design - Spotify OAuth with PocketBase

## Overview

Implement Spotify OAuth authentication with PocketBase 0.29 to create users linked to their Spotify accounts and protect API endpoints.

## Architecture

```
Frontend -> PocketBase Auth -> Spotify OAuth -> User Creation -> JWT Token -> Protected API
```

## Authentication Flow

### 1. Initial Setup
- Configure Spotify OAuth app in Spotify Developer Dashboard
- Set redirect URI: `https://yourdomain.com/auth/spotify/callback`
- Store Spotify client credentials in PocketBase settings

### 2. User Authentication Flow

#### Step 1: Initiate OAuth
```
GET /auth/spotify/login
```
- Redirect to Spotify OAuth URL with required scopes:
  - `user-read-email`
  - `playlist-read-private`
  - `playlist-modify-public`
  - `playlist-modify-private`

#### Step 2: OAuth Callback
```
GET /auth/spotify/callback?code=...&state=...
```
- Exchange authorization code for access/refresh tokens
- Fetch user profile from Spotify API
- Create or update user in PocketBase `users` collection

#### Step 3: User Creation/Linking
```go
type User struct {
    ID           string `json:"id"`
    Email        string `json:"email"`
    Name         string `json:"name"`
    SpotifyID    string `json:"spotify_id"`
    AccessToken  string `json:"access_token"`  // encrypted
    RefreshToken string `json:"refresh_token"` // encrypted
    TokenExpiry  time.Time `json:"token_expiry"`
    Created      time.Time `json:"created"`
    Updated      time.Time `json:"updated"`
}
```

#### Step 4: PocketBase Authentication
- Generate PocketBase auth token for user
- Return token to frontend with user info

## Implementation Components

### 1. Custom Auth Handlers
```go
// cmd/pb/auth.go
func setupSpotifyAuth(app *pocketbase.PocketBase) {
    app.OnServe().BindFunc(func(e *core.ServeEvent) error {
        e.Router.GET("/auth/spotify/login", spotifyLoginHandler)
        e.Router.GET("/auth/spotify/callback", spotifyCallbackHandler)
        return e.Next()
    })
}
```

### 2. Middleware for API Protection
```go
// Middleware to extract user from PocketBase token
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractBearerToken(r)
        user, err := validatePocketBaseToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 3. Token Management
- Store encrypted Spotify tokens in PocketBase
- Implement token refresh logic
- Handle token expiration gracefully

## Frontend Integration

### Login Flow
```javascript
// Redirect to login
window.location.href = '/auth/spotify/login';

// After successful auth, PocketBase token is available
const token = localStorage.getItem('pocketbase_token');
```

### API Calls
```javascript
// All API calls include PocketBase token
fetch('/api/base_playlist', {
    headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
    }
});
```

## Security Considerations

### Token Storage
- Encrypt Spotify tokens in database
- Use secure HTTP-only cookies for PocketBase tokens (alternative to localStorage)
- Implement token rotation

### API Protection
- All `/api/*` routes require authentication
- Rate limiting per user
- CORS configuration for frontend domain only

### Environment Variables
```env
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
SPOTIFY_REDIRECT_URI=https://yourdomain.com/auth/spotify/callback
ENCRYPTION_KEY=32_byte_key_for_token_encryption
```

## Database Schema Updates

### Users Collection
```sql
-- Add to existing users collection
ALTER TABLE users ADD COLUMN spotify_id TEXT UNIQUE;
ALTER TABLE users ADD COLUMN access_token TEXT; -- encrypted
ALTER TABLE users ADD COLUMN refresh_token TEXT; -- encrypted  
ALTER TABLE users ADD COLUMN token_expiry DATETIME;
```

## Implementation Priority

1. **Phase 1**: Basic Spotify OAuth flow
2. **Phase 2**: User creation/linking
3. **Phase 3**: API authentication middleware
4. **Phase 4**: Token encryption and refresh
5. **Phase 5**: Frontend integration

## Testing Strategy

- Unit tests for auth handlers
- Integration tests with Spotify OAuth (mock)
- E2E tests for complete auth flow
- Security testing for token handling