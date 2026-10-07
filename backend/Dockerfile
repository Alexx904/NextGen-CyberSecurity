# Stage 1: Build static Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 2: Minimal runtime environment with Nmap
FROM alpine:3.20

WORKDIR /app

# Install nmap, TLS certificates and tzdata
RUN apk add --no-cache \
    nmap \
    ca-certificates \
    tzdata

# Copy binary from builder stage
COPY --from=builder /app/server /app/server

# Expose default HTTP API port
EXPOSE 8080

ENTRYPOINT ["/app/server"]

