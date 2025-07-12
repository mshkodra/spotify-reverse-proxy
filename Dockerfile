# Use the official Go image as base
FROM golang:1.24.4-alpine

# Set environment variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Create and set working directory
WORKDIR /app

# Copy go.mod and go.sum before source files (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the app
RUN go build -o proxy ./cmd/proxy

# Expose the default port
EXPOSE 8080

# Run the app
CMD ["./proxy"]
