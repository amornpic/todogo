FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
# Verify the binary exists and is executable
RUN ls

# Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main /main

# Expose the application port
EXPOSE 3000

# Run the binary
CMD ["/main"]