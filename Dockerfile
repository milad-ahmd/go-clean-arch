# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o clean-arch-api ./cmd/api

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/clean-arch-api .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./clean-arch-api"]
