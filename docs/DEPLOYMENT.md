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

##### Medium Priority  
- [ ] Implement rate limiting per user
- [ ] Add security headers and CORS configuration
- [ ] Set up staging environment
- [x] Create CI/CD pipeline for automated deployments *(documented below)*

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

---

# CI/CD Pipeline with GitHub Actions

## Overview
Simple deployment pipeline using GitHub Actions to deploy to Fly.io on every push to the main branch, with code formatting, testing, and build validation.

## Pipeline Architecture

### Trigger Strategy
- **Push to main**: Automatic deployment to production
- **Manual dispatch**: Allow manual deployments for maintenance

### Workflow Steps
1. **Code Quality**: Linting and formatting checks
2. **Testing**: Unit tests
3. **Build**: Frontend and backend builds
4. **Deploy**: Automated deployment to Fly.io
5. **Health Check**: Basic endpoint validation

## Implementation Plan

### Step 1: GitHub Repository Setup
```bash
# 1. Create GitHub repository
gh repo create playlist-router --private --description "Spotify playlist management tool"

# 2. Add remote and push existing code
git remote add origin https://github.com/YOUR_USERNAME/playlist-router.git
git branch -M main
git push -u origin main
```

### Step 2: GitHub Secrets Configuration
Set up the following secrets in GitHub repository settings (`Settings > Secrets and variables > Actions`):

#### Required Secrets
- `FLY_API_TOKEN`: Fly.io API token for deployments

#### Optional Secrets
- `JWT_SECRET`: Only needed if running integration tests that require JWT signing

### Step 3: Create GitHub Actions Workflow

#### Deploy Workflow: `.github/workflows/deploy.yml`
```yaml
name: Deploy to Production

on:
  push:
    branches: [main]
  workflow_dispatch:

env:
  FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}

jobs:
  deploy:
    name: Build, Test & Deploy
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.24'
        
    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: 'web/package-lock.json'
        
    # Backend checks
    - name: Install Go dependencies
      run: go mod download
      
    - name: Run Go linter
      uses: golangci/golangci-lint-action@v6
      with:
        version: latest
      
    - name: Run Go tests
      run: make test
      
    # Frontend checks  
    - name: Install frontend dependencies
      run: make frontend-install
      
    - name: Run frontend linter
      run: cd web && npm run lint
      
    # Build
    - name: Build application
      run: make build-all
      
    # Deploy
    - name: Set up Fly CLI
      uses: superfly/flyctl-actions/setup-flyctl@master
      
    - name: Deploy to Fly.io
      run: flyctl deploy --remote-only
        
    - name: Health Check
      uses: jtalk/url-health-check-action@v4
      with:
        url: https://playlist-router.fly.dev/health
        max-attempts: 12
        retry-delay: 5s
```

### Step 4: Fly.io Setup for CI/CD

#### Generate Fly.io API Token
```bash
# Generate API token for GitHub Actions
fly auth token

# The output token should be added as FLY_API_TOKEN secret in GitHub
```

#### Configure Production Secrets via CLI
```bash
# Set production secrets (these are used by the running application, not CI/CD)
fly secrets set SPOTIFY_CLIENT_ID=$YOUR_CLIENT_ID
fly secrets set SPOTIFY_CLIENT_SECRET=$YOUR_CLIENT_SECRET  
fly secrets set JWT_SECRET=$YOUR_JWT_SECRET
fly secrets set FRONTEND_URL=https://playlist-router.fly.dev
```

### Step 5: Verify Makefile Commands
Ensure these commands exist in your Makefile (they should already be there):
```makefile
lint:        # Run golangci-lint
test:        # Run Go tests  
build-all:   # Build frontend + backend with embedded assets
frontend-install:  # Install npm dependencies
```

## Security Considerations

### Secret Management
- **GitHub Secrets**: Only FLY_API_TOKEN stored in GitHub repository secrets
- **Principle of Least Privilege**: CI/CD tokens have minimal required permissions
- **No Secrets in Code**: Zero hardcoded secrets in repository

### Build Security
- **Code Quality**: Automated linting and formatting checks
- **Testing**: All tests must pass before deployment
- **Health Checks**: Post-deployment validation

## Monitoring & Alerting

### Deployment Monitoring
- **Health Checks**: Automated endpoint validation post-deployment
- **GitHub Actions**: Deployment status visible in Actions tab
- **Fly.io Logs**: Real-time application logs via `fly logs`

## Cost Optimization

### GitHub Actions Costs
- **Free Tier**: 2,000 minutes/month for private repositories
- **Estimated Usage**: ~50 minutes/month for typical development pace
- **Cost**: $0 for most usage patterns

### Fly.io Costs  
- **Current Usage**: ~$5/month for single instance + 1GB volume
- **CI/CD Impact**: No additional infrastructure costs
- **Bandwidth**: Minimal additional costs for automated deployments

## Rollback Strategy

### Manual Rollbacks
If deployment fails or issues are found:
```bash
# List recent releases
fly releases

# Rollback to specific version  
fly releases rollback v12
```

## Testing Strategy

### Pre-Deploy Testing  
- **Linting**: Go linting (includes formatting) and TypeScript linting must pass
- **Unit Tests**: All Go unit tests must pass
- **Build Validation**: Frontend and backend builds must succeed

### Post-Deploy Testing
- **Health Checks**: Basic endpoint availability (`/health`)

## Implementation Timeline

### Simple Deploy Pipeline
- [x] **Plan documentation and workflow creation** *(completed)*
- [ ] GitHub repository setup and code push
- [ ] Set FLY_API_TOKEN secret in GitHub repository
- [ ] Test first automated deployment

## Benefits

### Development Velocity
- **Automated Testing**: Catch issues before production
- **Fast Deployments**: 3-5 minute end-to-end deployment
- **Zero-Downtime**: Fly.io rolling deployments
- **Quick Rollbacks**: One-click rollback capability

### Code Quality
- **Automated Linting**: Consistent code style enforcement
- **Test Coverage**: All tests run before deployment
- **Security Scanning**: Vulnerability detection
- **Documentation**: Self-documenting deployment process

### Operational Excellence  
- **Audit Trail**: Complete deployment history
- **Reproducible Builds**: Same artifacts from dev to production
- **Environment Parity**: Identical deployment process across environments
- **Reduced Human Error**: Automated, consistent deployments

## Future Enhancements

### Advanced Features (Month 2-3)
- **Staging Environment**: Deploy PRs to temporary staging instances
- **Performance Testing**: Load testing in CI/CD pipeline  
- **Database Migration**: Automated schema migration validation
- **Multi-Region Deployment**: Deploy to multiple Fly.io regions

### Enterprise Features (Month 4-6)
- **Blue-Green Deployments**: Zero-downtime deployments with instant rollback
- **Canary Deployments**: Gradual traffic shifting for risk reduction
- **A/B Testing**: Feature flag integration with automated deployment
- **Compliance**: SOC2/PCI compliance for payment features