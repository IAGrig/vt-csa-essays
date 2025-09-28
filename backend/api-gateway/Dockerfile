FROM golang:1.25.1-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum ./
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
#COPY essay-service/go.mod essay-service/go.sum ./essay-service/
COPY essay-service/go.mod ./essay-service/
COPY review-service/go.mod review-service/go.sum ./review-service/
COPY proto ./proto
COPY shared ./shared
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/

RUN go work sync

COPY api-gateway ./api-gateway

WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o service-binary ./cmd/server/main.go

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/api-gateway/service-binary .

ENTRYPOINT ["./service-binary"]
