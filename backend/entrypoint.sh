#!/bin/sh
set -e

# Wait for PostgreSQL to be ready
until pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER
do
  echo "Waiting for PostgreSQL to be available..."
  sleep 2
done

# Run migrations
echo "Running database migrations..."
goose -dir migrations postgres "user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB_NAME host=$POSTGRES_HOST port=$POSTGRES_PORT sslmode=disable" up

# Start application
exec "$@"
