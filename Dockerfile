# Mosychlos Go Application Dockerfile
FROM golang:alpine AS builder

WORKDIR /app

# No additional packages needed - all Go dependencies are standard modules

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o mosychlos ./cmd/mosychlos

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary
COPY --from=builder /app/mosychlos .

# Copy configuration files
COPY config/ ./config/

# Expose port if needed
EXPOSE 3000

CMD ["./mosychlos"]
