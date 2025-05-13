# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install goose and build tools
RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Production stage
FROM alpine:3.21
RUN apk add --no-cache postgresql-client
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/entrypoint.sh .
COPY --from=builder /app/.env .

RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./main"]
