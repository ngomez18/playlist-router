# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PlaylistSync is a Spotify playlist management tool that automatically distributes songs from "base" playlists to multiple themed "child" playlists based on configured rules. The application targets Spotify Premium subscribers who maintain multiple playlists.

## Planned Architecture

This is currently a planning-stage repository. When implementation begins, the architecture will be:

### Backend
- **Framework**: Go with PocketBase for rapid development
- **Database**: SQLite (via PocketBase)
- **Authentication**: JWT tokens with PocketBase auth + Spotify OAuth
- **API**: REST endpoints for playlist management and sync operations
- **Deployment**: Single Go binary on Fly.io serving both backend and frontend

### Frontend
- **Framework**: React 18 with TypeScript
- **UI Library**: Chakra UI (dark theme, responsive design)
- **State Management**: React Query for API state + Context for app state
- **Build Tool**: Vite
- **Styling**: Mobile-first, desktop-friendly responsive design

### Repository Structure (Planned)
- Monorepo with Go backend and React frontend
- Single deployment pipeline for integrated full-stack application
- Manual TypeScript/Go type definitions (future automation via PocketBase TypeGen)

## Key Data Models (From PRD)

### User
- Stores Spotify credentials, subscription tier, and usage tracking
- Contains monthly sync count for tier limits

### BasePlaylist
- Represents source playlists that users add songs to
- Links to Spotify playlist via spotify_playlist_id
- Users can have multiple base playlists

### ChildPlaylist
- Filtered playlists created from base playlists
- Contains comprehensive filtering rules based on Spotify audio features
- Supports exclusion lists for artists/songs

### SyncLog
- Tracks all sync operations and errors
- Provides audit trail for song distributions

## Core Functionality

### Spotify Integration
- OAuth authentication with Spotify
- Real-time playlist monitoring and song detection
- Audio feature analysis for filtering (energy, danceability, valence, etc.)
- Automatic playlist creation and song distribution

### Filtering System
- Supports all Spotify audio features as filter criteria
- Artist and song exclusion capabilities
- Pre-defined playlist templates
- Complex rule combinations

### Sync Engine
- Automatic detection of new songs in base playlists
- Real-time distribution to matching child playlists
- Manual sync trigger option
- Comprehensive error handling and retry logic

## Development Commands

**ALWAYS use the commands available in the Makefile instead of running Go commands directly:**

```bash
make help    # Show all available commands
make build   # Build the application
make test    # Run all tests  
make mocks   # Generate all mocks using go generate
make lint    # Run golangci-lint to check code quality
make fix     # Format and fix code issues
make deps    # Download and tidy dependencies
make dev     # Run with hot reload (air)
```

## Business Logic Considerations

### Subscription Tiers
- Free: 10 song distributions/month
- Basic ($0.99/month): Unlimited distributions, 2 base playlists, 5 children each
- Premium ($4.99/month): Unlimited everything

### Performance Requirements
- Page loads < 2 seconds
- API responses < 500ms
- Sync operations < 30 seconds for 50 songs
- Spotify API rate limit compliance (100 requests/minute)

### Security Requirements
- HTTPS everywhere
- Secure Spotify token storage
- Input validation and sanitization
- Rate limiting per user

## Development Notes
- Focus on Spotify Premium users as primary market
- Mobile-responsive design with multi-step wizards for complex forms
- Comprehensive error handling for Spotify API interactions
- Usage tracking and tier limit enforcement from day one
- Optimize for multiple base playlist scenarios (differentiator from competitors)

## Claude Instructions
### Development Guidelines
- Do what has been asked; nothing more, nothing less
- Use documentation for Pocketbase v0.29
- Maintain a clear separation of concerns: Repositories handle DB related operations. Services handle all business logic, and use repositories. Controllers handle request logic. Parsing, validation, responses, etc. They use the services
- Logging and proper error handling should always be taken into account
- Don't add unnecessary comments to the code. Only annotate important functions, types, or important pieces of code
- **Don't implement unit tests for large pieces of code until I've reviewed it and tell you to implement them**
- Unit tests are important. Opt for table-driven tests when possible. Dependency injection is prefered so test code can be easily generated. Use testify by initializing an assert instance for every test and using that for assertions
- When asked for your thoughts on something, try to be as critical as possible. Don't blindly accept suggestions, always weigh the pros and cons
- When asked to plan large features or changes, generate a markdown file in the docs/ directory with the final result of the discussion
- Be concise. Provide only the necessary information
- **I don't like unnecessary comments. Only really difficult to understand parts of our code should be commented, no need to add comments explaining very simple pieces of code or things that are properly named and self-explanatory**

### MANDATORY: Use Makefile Commands
**ALWAYS use Makefile commands instead of direct Go commands:**
- `make test` instead of `go test`
- `make build` instead of `go build` 
- `make mocks` instead of `go generate`
- `make lint` instead of `golangci-lint run`
- `make fix` instead of `gofmt` or `goimports`
```