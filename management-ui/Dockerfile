FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for cloning repositories
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod ./
# COPY go.sum ./ # Uncomment if you add dependencies

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o management-ui .

# Final stage
FROM alpine:latest

# Install docker CLI, compose plugin, and git
RUN apk add --no-cache docker-cli docker-cli-compose docker-cli-compose git

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/management-ui .
COPY --from=builder /app/web ./web

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./management-ui"]
