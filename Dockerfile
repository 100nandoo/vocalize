# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o vocalize .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/vocalize .

# Copy web assets
COPY --from=builder /app/web ./web

# Copy .env if it exists (optional)
COPY .env .env.example* ./

# Expose port
EXPOSE 8080

# Run the application
CMD ["./vocalize", "serve"]
