FROM golang:1.24-alpine

RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.11.0

WORKDIR /app
COPY . .

CMD ["sh", "-c", "goose -dir . postgres \
  \"user=$POSTGRES_USER \
   password=$POSTGRES_PASSWORD \
   dbname=$POSTGRES_DB_NAME \
   host=$POSTGRES_HOST \
   port=$POSTGRES_PORT \
   sslmode=disable\" up"]
