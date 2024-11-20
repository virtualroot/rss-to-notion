FROM golang:1.23.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o rss-to-notion main.go

# Use distroless as runtime base image
FROM gcr.io/distroless/static:nonroot

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /app/rss-to-notion /app/rss-to-notion

# Use non-root user for security
USER nonroot:nonroot

ENTRYPOINT ["/app/rss-to-notion"]
CMD ["help"]
