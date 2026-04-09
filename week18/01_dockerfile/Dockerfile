# ========================================
# Multi-Stage Dockerfile for a Go Service
# ========================================
#
# Multi-stage builds are the standard for Go Docker images.
# Stage 1 (builder): compiles the Go binary with all build tools
# Stage 2 (runtime): copies ONLY the binary into a tiny image
#
# This produces a container that is ~10-15 MB instead of ~800 MB!
#
# Build: docker build -t go-lesson-service .
# Run:   docker run -p 8080:8080 -e ENVIRONMENT=production go-lesson-service

# ========================================
# STAGE 1: Builder
# ========================================
# Use the official Go image as the build environment.
# This image is ~800 MB but we only use it temporarily.
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies)
# and tzdata (for timezone support in the binary)
RUN apk add --no-cache git ca-certificates tzdata

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod first (for better Docker layer caching).
# If go.mod hasn't changed, Docker reuses the cached layer
# and skips "go mod download" — this speeds up rebuilds.
COPY go.mod ./
RUN go mod download

# Now copy the rest of the source code
COPY . .

# Build the binary with optimizations:
#   CGO_ENABLED=0  — pure Go binary (no C dependencies)
#   -ldflags="-s -w" — strip debug info (smaller binary)
#   -o /app/server — output path
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server .

# ========================================
# STAGE 2: Runtime
# ========================================
# Use a minimal base image. Options:
#   scratch  — empty image (smallest, ~0 MB, no shell/debug tools)
#   alpine   — minimal Linux (~5 MB, has shell for debugging)
#   distroless — Google's minimal images (no shell, but has CA certs)
#
# We use alpine here because it's great for learning (you can exec into it).
# In production, consider scratch or distroless for security.
FROM alpine:3.21

# Install CA certificates so the service can make HTTPS calls.
# If using scratch, you'd COPY them from the builder stage.
RUN apk add --no-cache ca-certificates

# Create a non-root user for security.
# NEVER run containers as root in production!
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy ONLY the compiled binary from the builder stage.
# This is what makes multi-stage builds powerful — the final
# image doesn't contain Go, source code, or build tools.
COPY --from=builder /app/server .

# Copy timezone data from the builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Switch to the non-root user
USER appuser

# Document which port the service uses.
# This doesn't actually publish the port — it's documentation.
# You still need -p 8080:8080 when running.
EXPOSE 8080

# Set default environment variables.
# These can be overridden at runtime with -e or docker-compose.
ENV PORT=8080
ENV ENVIRONMENT=production

# Health check — Docker will monitor this endpoint
# and mark the container as unhealthy if it fails.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# The command to run when the container starts.
# Use the exec form (JSON array) — it runs the binary directly
# without a shell, which handles signals correctly.
CMD ["/app/server"]
