# Stage 1: Build Environment
FROM golang:1.24.1 AS builder

WORKDIR /usr/src/app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy project
COPY . .

# Build project
RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/app ./cmd/pvz-service

# Stage 2: Runtime environment
FROM alpine:3.21

WORKDIR /

# Copy migrations
COPY --from=builder /usr/src/app/internal/repository/migrations/ /migrations

# Copy configs and app file
COPY --from=builder /usr/src/app/config/ /config
COPY --from=builder /usr/local/bin/app /app

CMD ["/app"]