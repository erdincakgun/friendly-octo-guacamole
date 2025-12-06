# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy source code
COPY main.go .

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" to strip debug info and reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mock-service main.go

# Final stage - minimal image
FROM scratch

# Copy the binary from builder
COPY --from=builder /app/mock-service /mock-service

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/mock-service"]


