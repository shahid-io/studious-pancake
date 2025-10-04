# Start from the official Golang image

FROM golang:1.24-alpine AS builder
WORKDIR /workspace

COPY . .
WORKDIR /workspace/services/auth-service
RUN go mod download
RUN go build -o auth-service main.go

# Build the Go app
RUN go build -o auth-service main.go


# Copy the entire project for Go module resolution
WORKDIR /workspace
COPY ../../ .
WORKDIR /workspace/services/auth-service
RUN go mod download
RUN go build -o auth-service main.go

# Build the Go app
RUN go build -o auth-service main.go

# Use a minimal image for running

FROM alpine:latest
WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /workspace/services/auth-service/auth-service ./auth-service

EXPOSE 8080

CMD ["./auth-service"]
