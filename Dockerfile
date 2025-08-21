# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23 AS builder
WORKDIR /app

# Download dependencies first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /pepo-server ./cmd/server

# Run stage
FROM gcr.io/distroless/base-debian12

# Set working directory and copy binary
WORKDIR /
COPY --from=builder /pepo-server /pepo-server

# Expose service port and run
EXPOSE 8080
ENV PORT=8080
CMD ["/pepo-server"]
