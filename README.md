# PlaylistRouter

PlaylistRouter is a Spotify playlist management tool that automates song distribution from source playlists to destination playlists based on configurable audio feature rules.

## Technology Stack

**Backend**
*   **Language**: Go (v1.24+)
*   **Framework**: PocketBase (SQLite, Authentication, API)
*   **Architecture**: Monolithic binary serving API and static assets

**Frontend**
*   **Framework**: React 18
*   **Language**: TypeScript
*   **Build Tool**: Vite
*   **UI Library**: DaisyUI

**Infrastructure**
*   **Deployment**: Containerized application deployed to Fly.io

## Dependencies

*   Go 1.24 or higher
*   Node.js 18 or higher
*   npm

## Project Structure

*   `cmd/`: Application entry points
*   `internal/`: Backend business logic and services
*   `web/`: React frontend application
*   `docs/`: Project documentation and design specifications
*   `pb_data/`: PocketBase data directory (local development)

## Usage

The project uses a `Makefile` for standard development operations.

### Installation

```bash
# Install frontend dependencies
make frontend-install

# Download Go modules
make deps
```

### Development

To run the full application in development mode (Backend + Frontend with hot reload):

```bash
make dev-full
```

*   **Backend API**: http://localhost:8090
*   **Frontend**: http://localhost:5173
*   **Admin UI**: http://localhost:8090/_/

### Build

To build the production binary with embedded frontend assets:

```bash
make build-all
```

### Testing

```bash
make test
```

## Documentation

*   [Product Requirements](docs/PRD.md)
*   [Deployment Guide](docs/DEPLOYMENT.md)
*   [API Design](docs/API-DESIGN.md)
*   [Database Schema](docs/DB-SCHEMA.md)
*   [Authentication](docs/AUTH-DESIGN.md)
*   [Sync Logic](docs/SYNC-DESIGN.md)
