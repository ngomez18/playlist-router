# Multi-stage Docker build for PlaylistRouter
##################################################
# Stage 1: Build frontend with Node.js
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci

# Copy frontend source and build
COPY web/ ./
RUN npm run build

##################################################
# Stage 2: Build Go application
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

# Copy Go source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Copy built frontend directly to the embedded location
RUN mkdir -p internal/static/dist
COPY --from=frontend-builder /app/web/dist/ ./internal/static/dist/

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o playlist-router ./cmd/pb

##################################################
# Stage 3: Final runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -s /bin/sh appuser
RUN mkdir -p /app /data && chown -R appuser:appuser /app /data

WORKDIR /app

# Copy the binary from builder stage
COPY --from=backend-builder /app/playlist-router .
RUN chown appuser:appuser playlist-router && chmod +x playlist-router

# Change to non-root user
USER appuser

ENV PORT=8080
EXPOSE $PORT

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:$PORT/health || exit 1


CMD sh -c "./playlist-router serve --http=0.0.0.0:$PORT --dir=/data"