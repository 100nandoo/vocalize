# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    libopus-dev \
    libopusfile-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGo enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o inti .

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libopus0 \
    libopusfile0 \
    wget \
    tesseract-ocr \
    tesseract-ocr-eng \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/inti .

# Copy web assets
COPY --from=builder /app/web ./web

# Expose port
EXPOSE 8282

# Run the application
CMD ["./inti", "serve"]
