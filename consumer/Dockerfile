# Build stage
FROM golang:1.21 as builder

WORKDIR /build

# Install required dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    pkg-config \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -o consumer .

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    librdkafka1 \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary from builder
COPY --from=builder /build/consumer /usr/local/bin/consumer

# Run the application
CMD ["consumer"]
