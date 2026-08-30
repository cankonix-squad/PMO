# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/bin/cankora ./cmd/api/main.go

# ---- Runtime image ----
FROM alpine:3.19

WORKDIR /app

# Add timezone data and ca-certs for HTTPS/SMTP
RUN apk add --no-cache ca-certificates tzdata

# Copy binary and migrations
COPY --from=builder /app/bin/cankora ./cankora
COPY --from=builder /app/migrations ./migrations

# Create uploads directory
RUN mkdir -p ./uploads

EXPOSE 8080

ENTRYPOINT ["./cankora"]
