# Deployment Guide - Fly.io

## Overview
Deploy PlaylistRouter as a single Docker container on Fly.io, serving both the Go backend API and React frontend as static files.

## Architecture
- **Single Binary**: Go application serving both API and frontend
- **Static Assets**: React build embedded in Go binary
- **Database**: SQLite with PocketBase (persistent volume)
- **Security**: API endpoints protected by authentication, not publicly accessible

## Pre-Deployment Considerations

### 1. Environment Configuration
- [ ] Production environment variables
- [ ] Spotify OAuth redirect URLs for production domain
- [ ] Database file persistence
- [ ] Secret management (Spotify client credentials, JWT secrets)

### 2. Build Strategy
- [ ] Multi-stage Docker build (Node.js for frontend, Go for backend)
- [ ] Frontend build integration with Go binary
- [ ] Static asset embedding

### 3. Security
- [ ] CORS configuration for production domain
- [ ] Rate limiting implementation
- [ ] API authentication enforcement
- [ ] HTTPS enforcement

### 4. Database Persistence
- [ ] Fly.io volume for SQLite database
- [ ] Database initialization on first deployment
- [ ] Backup strategy

## Deployment Steps

### Step 1: Create Dockerfile
```dockerfile
# Multi-stage build
FROM node:18-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o playlist-router ./cmd/pb

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=backend-builder /app/playlist-router .
EXPOSE 8090
CMD ["./playlist-router", "serve", "--http=0.0.0.0:8090"]
```

### Step 2: Configure fly.toml
```toml
app = "playlist-router"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8090"
  PB_DATA_DIR = "/data"

[[services]]
  internal_port = 8090
  protocol = "tcp"

  [[services.ports]]
    handlers = ["http"]
    port = 80
    force_https = true

  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443

[mounts]
  source = "data"
  destination = "/data"
```

### Step 3: Environment Setup
Required secrets to set via `fly secrets set`:
- `SPOTIFY_CLIENT_ID`
- `SPOTIFY_CLIENT_SECRET`
- `JWT_SECRET`
- `FRONTEND_URL` (production domain)

### Step 4: Database Volume
```bash
fly volumes create data --size 1 --region iad
```
*Note: 1GB volume is sufficient for small friend group usage*

### Step 5: Frontend Build Integration
Update `internal/static/static.go` to embed built frontend:
```go
//go:embed web/dist/*
var frontendFS embed.FS

func GetFrontendFS() (fs.FS, error) {
    return fs.Sub(frontendFS, "web/dist")
}
```

## Security Considerations

### API Protection
1. **Authentication Required**: All API endpoints already require auth middleware
2. **CORS Configuration**: Restrict to production domain only
3. **Rate Limiting**: Implement per-user rate limiting
4. **Input Validation**: Ensure all inputs are validated and sanitized

### Recommended CORS Configuration
```go
// Add to main.go
func setupCORS(handler http.Handler, allowedOrigin string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        handler.ServeHTTP(w, r)
    })
}
```

## Production Configuration Changes

### 1. Spotify OAuth Redirect URLs
Update Spotify app settings to include:
- `https://playlist-router.fly.dev/auth/spotify/callback`

### 2. Frontend Configuration
Update frontend API base URL for production:
```typescript
const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? '' // Same origin
  : 'http://localhost:8090'
```

### 3. PocketBase Configuration
Ensure PocketBase serves from correct data directory:
```go
app := pocketbase.NewWithConfig(pocketbase.Config{
    DataDir: os.Getenv("PB_DATA_DIR"),
})
```

## Monitoring & Observability

### Logging Strategy
- [ ] Structured logging with slog
- [ ] Log aggregation via fly.io logs
- [ ] Error tracking and alerting

### Health Checks
Add health check endpoint:
```go
e.Router.GET("/health", apis.WrapStdHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})))
```

## Deployment Commands

### Initial Deployment
```bash
# 1. Initialize fly app
fly apps create playlist-router

# 2. Set secrets
fly secrets set SPOTIFY_CLIENT_ID=your_client_id
fly secrets set SPOTIFY_CLIENT_SECRET=your_client_secret
fly secrets set JWT_SECRET=your_jwt_secret
fly secrets set FRONTEND_URL=https://playlist-router.fly.dev

# 3. Create volume
fly volumes create data --size 1 --region iad

# 4. Deploy
fly deploy
```

### Subsequent Deployments
```bash
fly deploy
```

## Rollback Strategy
```bash
# View releases
fly releases

# Rollback to previous version
fly releases rollback <version>
```

## Database Backup
```bash
# Access the app instance
fly ssh console

# Backup database
cp /data/data.db /data/backup-$(date +%Y%m%d).db
```

## Performance Optimization

### Static Asset Optimization
- [ ] Gzip compression for static files
- [ ] Cache headers for static assets
- [ ] CDN consideration for static assets

### Go Application
- [ ] Connection pooling
- [ ] Request timeout configuration
- [ ] Memory optimization

## Configuration Decisions

1. **Domain**: Using default `.fly.dev` domain (playlist-router.fly.dev)

2. **Database Size**: Small dataset (friends only) - 1GB volume sufficient

3. **Regional Deployment**: US-East region for South American users

4. **Monitoring**: Fly.io built-in logging only for now

5. **Scaling**: Single instance sufficient initially

6. **Backup Frequency**: Weekly manual backups to start

7. **API Rate Limits**: PocketBase built-in rate limiting (5 req/sec per user)

8. **Environment Separation**: Production only for now

## Next Steps

1. Set up fly.io account and CLI
2. Configure Spotify OAuth for playlist-router.fly.dev domain  
3. Create Dockerfile and fly.toml
4. Update static file embedding for frontend
5. Configure PocketBase rate limiting
6. Test build process locally
7. Deploy to fly.io
8. Set up weekly backup procedures

## Estimated Timeline
- Initial setup and configuration: 2-3 hours
- Testing and debugging: 1-2 hours
- Production deployment: 1 hour
- Monitoring setup: 1 hour

**Total: 5-7 hours**