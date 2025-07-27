# Authentication Design - Spotify OAuth with PocketBase

## Overview

Implement Spotify OAuth authentication with PocketBase 0.29 to create users linked to their Spotify accounts and protect API endpoints.

## Architecture

```
Frontend -> PocketBase Auth -> Spotify OAuth -> User Creation -> JWT Token -> Protected API
```

## Authentication Flow

### 1. Initial Setup c
- Configure Spotify OAuth app in Spotify Developer Dashboard
- Set redirect URI: `http://127.0.0.1:8090/auth/spotify/callback` (dev)
- Store Spotify client credentials in environment variables (.env file)

### 2. User Authentication Flow

#### Step 1: Initiate OAuth ✅ IMPLEMENTED
```
GET /auth/spotify/login
```
- Redirect to Spotify OAuth URL with required scopes:
  - `user-read-email`
  - `playlist-read-private`
  - `playlist-modify-public`
  - `playlist-modify-private`
- Generate random state for CSRF protection

#### Step 2: OAuth Callback ✅ IMPLEMENTED
```
GET /auth/spotify/callback?code=...&state=...
```
- ✅ Exchange authorization code for access/refresh tokens
- ✅ Fetch user profile from Spotify API  
- ✅ Create or update user in PocketBase (returns placeholder data currently)

#### Step 3: User Creation/Linking ❌ TODO
- Need to implement actual user creation in PocketBase
- Store user data from Spotify profile  
- Link to spotify_integrations table for tokens
- Use separate tables: `users` and `spotify_integrations` (as per DB-SCHEMA.md)

#### Step 4: PocketBase Authentication ❌ TODO
- Generate PocketBase auth token for user
- Return token to frontend with user info
- Frontend needs to store and use tokens for API calls

## Implementation Components

### 1. Custom Auth Handlers ✅ IMPLEMENTED
```go
// cmd/pb/main.go - initAppRoutes()
auth := e.Router.Group("/auth")
auth.GET("/spotify/login", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyLogin)))
auth.GET("/spotify/callback", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyCallback)))
```
- ✅ AuthController with SpotifyLogin and SpotifyCallback methods
- ✅ AuthService with business logic  
- ✅ SpotifyClient for API interactions

### 2. Middleware for API Protection ❌ TODO
- Need to implement auth middleware for `/api/*` routes
- Extract and validate PocketBase JWT tokens  
- Add user context to requests
- Currently API endpoints use placeholder user IDs

### 3. Token Management ❌ TODO
- ❌ Store encrypted Spotify tokens in PocketBase
- ❌ Implement token refresh logic
- ❌ Handle token expiration gracefully
- Currently tokens are fetched but not persisted

## Frontend Integration

### Login Flow ⚠️ PARTIALLY IMPLEMENTED
```javascript
// ✅ Redirect to login (with environment variable support)
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ''
window.location.href = `${apiBaseUrl}/auth/spotify/login`;

// ❌ After successful auth, PocketBase token handling needed
// TODO: Handle callback response and store tokens
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