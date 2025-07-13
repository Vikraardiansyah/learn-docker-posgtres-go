# Dockerfile
# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum from the build context (which is '.')
COPY go.mod go.sum ./

RUN go mod download

# Copy all application code from the build context (which is '.')
# This will copy main.go and any other source files directly into /app
COPY . .

# Build the Go application
# The path to main.go inside the container is now ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./main.go


# Stage 2: Create a minimal image for running the app
FROM alpine:latest

# Install necessary certificates for HTTPS if your Go app makes external HTTP requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /app/app .

# Expose the port your Go application listens on
EXPOSE 8080

# Command to run the application when the container starts
CMD ["./app"]