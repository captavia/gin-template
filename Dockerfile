# ==========================================
# Stage 1: Build the Go binary using CGO and Makefile
# ==========================================
FROM golang:alpine AS builder

# Install build tools necessary for CGO and Makefile (make, gcc, libc-dev, etc.)
RUN apk add --no-cache build-base git

WORKDIR /src

# Copy dependency files first to utilize Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go binary using Makefile targets

RUN make all

# ==========================================
# Stage 2: Create a minimal runtime image
# ==========================================
FROM alpine:latest

# Install runtime dependencies:
# - ca-certificates: Required for HTTPS requests (e.g. MinIO, external APIs)
# - tzdata: Required for time/timezone operations
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user and group for security best practices
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the compiled binary from the builder stage (Makefile output is in ./build/api)
COPY --from=builder --chown=appuser:appgroup /src/build/api /app/api

# Note: The application expects config.toml at runtime.
# You can mount this config file to /app/config.toml, 
# or copy it in if it is present during image build.

# Run the binary as the non-root user
USER appuser

# Expose the API port (usually 8080 or configurable)
EXPOSE 8080

# Run the API server
ENTRYPOINT ["/app/api"]
