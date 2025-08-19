# Deployment Guide - Fly.io

## Overview
Deploy PlaylistRouter as a single Docker container on Fly.io, serving both the Go backend API and React frontend as static files.

**✅ Status: DEPLOYED**  
**🌐 Live URL:** https://playlist-router.fly.dev  
**📊 Health Check:** https://playlist-router.fly.dev/health

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
# Multi-stage Docker build for PlaylistRouter
# Stage 1: Build frontend with Node.js
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files and install dependencies
COPY web/package*.json ./
RUN npm ci

# Copy frontend source and build for production
COPY web/ ./
RUN npm run build

# Stage 2: Build Go application
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy Go modules files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy only Go source code (exclude web/, docs/, etc.)
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Copy built frontend directly to the embedded location
RUN mkdir -p internal/static/dist
COPY --from=frontend-builder /app/web/dist/ ./internal/static/dist/

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o playlist-router ./cmd/pb

# Stage 3: Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Create app and data directories with proper permissions
RUN mkdir -p /app /data && chown -R appuser:appuser /app /data

WORKDIR /app

# Copy the binary from builder stage
COPY --from=backend-builder /app/playlist-router .
RUN chown appuser:appuser playlist-router && chmod +x playlist-router

# Change to non-root user
USER appuser

# Set environment variables
ENV PORT=8080

# Expose port (will be overridden by PORT env var)
EXPOSE $PORT

# Health check (uses localhost since we're inside the container)
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:$PORT/health || exit 1

# Start the application using PORT environment variable and data directory
CMD sh -c "./playlist-router serve --http=0.0.0.0:$PORT --dir=/data"
```

### Step 2: Configure fly.toml
```toml
app = "playlist-router"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8080"
  APP_ENV = "production"
  LOG_LEVEL = "info"

[[services]]
  internal_port = 8080
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

[[vm]]
  memory = 512
  cpu_kind = "shared"
  cpus = 1
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
//go:embed all:dist
var embeddedFiles embed.FS

func GetFrontendFS() (fs.FS, error) {
    return fs.Sub(embeddedFiles, "dist")
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
Ensure PocketBase serves from correct data directory using explicit flag:
```dockerfile
# In Dockerfile CMD
CMD sh -c "./playlist-router serve --http=0.0.0.0:$PORT --dir=/data"
```

This ensures PocketBase uses the mounted volume at `/data` instead of creating `pb_data/` in the container.

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

## Deployment Status

### ✅ Completed
- [x] Multi-stage Docker build (Node.js + Go)
- [x] fly.io app created and configured
- [x] Persistent volume for SQLite database (1GB)
- [x] Production secrets management
- [x] HTTPS and custom domain setup
- [x] Static asset serving with embedded frontend
- [x] Health check endpoint (`/health`)
- [x] Spotify OAuth configured for production domain

### 🔄 Next Steps (Operational Improvements)

#### High Priority
- [ ] Set up database backup automation
- [ ] Configure structured logging with levels
- [ ] Add error monitoring and alerting
- [ ] Document rollback procedures

#### Medium Priority  
- [ ] Implement rate limiting per user
- [ ] Add security headers and CORS configuration
- [ ] Set up staging environment
- [ ] Create CI/CD pipeline for automated deployments

#### Low Priority
- [ ] Add metrics collection and dashboards
- [ ] Configure auto-scaling policies
- [ ] Implement blue-green deployment strategy

## Deployment Timeline (Completed)

### Initial Deployment ✅
- **Setup and configuration**: ~3 hours
- **Docker multi-stage build**: ~1 hour  
- **Asset serving fixes**: ~2 hours
- **Testing and debugging**: ~2 hours
- **Production deployment**: ~1 hour

**Total Time Invested: ~9 hours**

### Lessons Learned
1. **Frontend environment variables**: Vite's handling of `VITE_API_BASE_URL` can break builds - be careful with empty strings
2. **Static asset serving**: Docker `COPY` with `/*` doesn't copy subdirectories - use `/` instead  
3. **Environment separation**: `.dockerignore` is cleaner than manual file deletion for excluding dev configs
4. **Port standardization**: Use 8080 for containers (industry standard) vs 8090 for local dev
5. **PocketBase data directory**: Use explicit `--dir=/data` flag instead of `PB_DATA_DIR` env var for clarity and proper volume mounting

### Future Deployments
With the current setup, future deployments are simple:
```bash
fly deploy  # ~2-3 minutes
```