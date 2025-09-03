# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.11.0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Production stage
FROM alpine:3.21
RUN apk add --no-cache postgresql-client
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

RUN chmod +x ./main

ENTRYPOINT ["./main"]
CMD []
