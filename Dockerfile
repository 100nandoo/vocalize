# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies (CGo requires build-base and opus-dev)
RUN apk add --no-cache git build-base opus-dev opusfile-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGo enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o vocalize .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates opus opusfile

# Copy binary from builder
COPY --from=builder /app/vocalize .

# Copy web assets
COPY --from=builder /app/web ./web

# Copy config files if they exist
COPY .env.example .env* ./

# Expose port
EXPOSE 8080

# Run the application
CMD ["./vocalize", "serve"]
