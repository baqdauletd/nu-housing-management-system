# ============================
# Stage 1 — Build the Go binary
# ============================
FROM golang:1.25-alpine AS builder

# Required for CGO-disabled static builds
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install git for go get
RUN apk add --no-cache git

# Create working directory
WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the backend binary
RUN go build -o server ./cmd/server/main.go


# ============================
# Stage 2 — Run the Go binary
# ============================
FROM alpine:3.18

# Install certificates for HTTPS (required for MinIO, Postgres, etc.)
RUN apk --no-cache add ca-certificates

# Create work directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/server .

# environment file if needed (optional)
COPY .env .env

EXPOSE 8080
CMD ["./server"]
