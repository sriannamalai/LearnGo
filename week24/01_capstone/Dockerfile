# ========================================
# TaskFlow Multi-Stage Docker Build
# ========================================
# Multi-stage builds create minimal production images by separating
# the build environment from the runtime environment. This demonstrates:
#   - Week 19-20: Docker patterns for Go applications
#
# Stage 1 (builder): Full Go toolchain, compiles the binary
# Stage 2 (runtime): Minimal base image, only the compiled binary
#
# The result is a ~15MB image instead of ~1GB with the full toolchain.
#
# Build:
#   docker build -t taskflow .
#
# Run:
#   docker run -p 8080:8080 -p 9090:9090 taskflow

# ========================================
# Stage 1: Build
# ========================================
FROM golang:1.25-alpine AS builder

# Install build dependencies
# - gcc and musl-dev are needed for CGO (some packages require it)
# - ca-certificates for HTTPS in the final image
RUN apk add --no-cache gcc musl-dev ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first (Docker layer caching optimization)
# If go.mod and go.sum haven't changed, Docker reuses the cached
# module download layer — significantly speeding up rebuilds.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with optimizations:
#   CGO_ENABLED=0: Static binary (no C library dependency)
#   -trimpath: Remove file system paths from binary (security)
#   -ldflags: Set version info and strip debug symbols
#     -s: Strip symbol table
#     -w: Strip DWARF debugging information
#     -X: Set version variables at build time
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X github.com/sri/learngo/week24/01_capstone/internal/model.Version=1.0.0" \
    -o /app/taskflow \
    ./01_capstone/

# ========================================
# Stage 2: Runtime
# ========================================
# Use a minimal base image. Options:
#   - scratch: Zero base image (smallest, but no shell or tools)
#   - alpine: ~5MB, includes shell and package manager
#   - distroless: Google's minimal images, no shell
#
# We use alpine for the educational example because it allows
# exec-ing into the container for debugging. In production,
# consider distroless for the smallest attack surface.
FROM alpine:3.20

# Install runtime dependencies only
# - ca-certificates: Required for HTTPS calls to external services
# - tzdata: Timezone data for time.LoadLocation()
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user for security
# Running as root in containers is a security risk — if the
# container is compromised, the attacker has root access.
RUN addgroup -S taskflow && adduser -S taskflow -G taskflow

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/taskflow .

# Copy configuration and migration files
COPY --from=builder /app/01_capstone/config.yaml .
COPY --from=builder /app/01_capstone/migrations ./migrations/

# Change ownership to the non-root user
RUN chown -R taskflow:taskflow /app

# Switch to non-root user
USER taskflow

# Expose the service ports
# EXPOSE doesn't publish ports — it documents which ports the
# container listens on. Use -p to actually publish ports.
EXPOSE 8080 9090

# Health check — Docker periodically runs this to verify the
# container is healthy. If it fails, Docker can restart it.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
# Use exec form (JSON array) instead of shell form to ensure
# the process receives signals correctly for graceful shutdown.
ENTRYPOINT ["./taskflow"]
CMD ["serve"]
