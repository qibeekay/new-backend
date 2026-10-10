# Multi-stage Dockerfile for Portfolio API
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# Final minimal runtime image
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/api /app/bin/api
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/docs /app/docs

# Local uploads directory (fallback)
RUN mkdir -p /app/uploads && chmod 755 /app/uploads

EXPOSE 8080

ENTRYPOINT ["/app/bin/api"]
