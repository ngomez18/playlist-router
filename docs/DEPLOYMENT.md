# Deployment Reference

**Status:** ✅ Deployed to Fly.io  
**URL:** https://playlist-router.fly.dev

## Architecture

PlaylistRouter is deployed as a **single Docker container** on Fly.io.
- **Backend**: Go application (PocketBase framework).
- **Frontend**: React SPA (Vite) built and embedded into the Go binary as static assets.
- **Database**: SQLite, persisted via a Fly.io Volume (`/data`).

## Environment Variables Strategy

### Backend (Runtime Secrets)
Backend environment variables are injected at runtime.
- **Source**: Fly.io Secrets.
- **Management**: `fly secrets set KEY=VALUE`
- **Key Variables**:
    - `SPOTIFY_CLIENT_ID` / `SPOTIFY_CLIENT_SECRET`: Spotify OAuth.
    - `JWT_SECRET`: For signing auth tokens.
    - `FRONTEND_URL`: Production domain for CORS/Redirects.

### Frontend (Build-time Configuration)
Frontend environment variables are **baked into the static files** during the Docker build.
- **Source**: GitHub Repository Secrets.
- **Process**:
    1. GitHub Actions workflow extracts secrets.
    2. Writes them to `web/.env.production`.
    3. Docker build copies this file.
    4. `npm run build` replaces variables in the JS bundle.
- **Key Variables**:
    - `VITE_FULLSTORY_ORG_ID`: Analytics.
    - `VITE_API_BASE_URL`: Set to empty string `''` for production (same-origin).

## Deployment Process

### Automated (CI/CD)
We use **GitHub Actions** for continuous deployment.
- **Trigger**: Push to `main` branch.
- **Workflow**: `.github/workflows/deploy.yml`
- **Steps**:
    1. **Test**: Runs Go unit tests and linters.
    2. **Config**: Creates `web/.env.production` from GitHub Secrets.
    3. **Build**: Runs `make build-all` (Frontend build -> Go embed -> Go build).
    4. **Deploy**: Uses `flyctl deploy --remote-only` to build and deploy the image on Fly.io builders.

### Manual Deployment
If needed, you can deploy manually from your local machine (ensure you have `flyctl` installed).

## Infrastructure Configuration

### Dockerfile
- **Multi-stage build**:
    1. `frontend-builder`: Node.js image to build React app.
    2. `backend-builder`: Go image to build binary with embedded assets.
    3. `runtime`: Alpine image, runs the binary.
- **Persistence**: Mounts `/data` volume for SQLite.

### fly.toml
- **App Name**: `playlist-router`
- **Region**: `iad` (US East)
- **Volume**: Mounts `data` volume to `/data`.
- **Port**: Exposes 8080 (mapped to 80/443).

## Operational Commands

```bash
# View logs
fly logs

# SSH into container
fly ssh console

# Backup database (manual)
fly ssh console -C "cp /data/data.db /data/backup-$(date +%Y%m%d).db"
```