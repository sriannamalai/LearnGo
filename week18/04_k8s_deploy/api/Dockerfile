# ========================================
# Dockerfile: API Gateway Service
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
    go build -ldflags="-s -w" -o /app/api-gateway .

# Stage 2: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/api-gateway .
USER appuser
EXPOSE 8080
ENV PORT=8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1
CMD ["/app/api-gateway"]
