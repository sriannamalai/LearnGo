# ========================================
# Dockerfile: Worker Service
# ========================================
# Multi-stage build optimized for production.

# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/worker .

# Stage 2: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/worker .
USER appuser
EXPOSE 8081
ENV PORT=8081
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8081/health || exit 1
CMD ["/app/worker"]
