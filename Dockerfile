# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary system dependencies (if any)
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# We point to cmd/server.go as the entry point based on your file structure
RUN go build -o chatapp cmd/server.go

# Stage 2: Create a small image for running the app
FROM alpine:latest

WORKDIR /root/

# Install CA certificates (required for HTTPS requests if your app makes any)
RUN apk --no-cache add ca-certificates
    
# Copy the binary from the builder stage
COPY --from=builder /app/chatapp .

# Copy static resources (Templates are required by your routes)
COPY --from=builder /app/templates ./templates

# Create the uploads directory (required by your upload handler)
RUN mkdir -p uploads

# Copy the .env file (optional, though environment variables in docker-compose are preferred)
# COPY .env .

# Expose the port defined in your .env/config
EXPOSE 8080

# Command to run the executable
CMD ["./chatapp"]