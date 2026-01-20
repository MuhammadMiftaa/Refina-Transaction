#!/bin/sh

set -e

# Ambil env variables yang di-inject (3 variabel teratas)
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
GOOSE_DBSTRING=${GOOSE_DBSTRING}

# Ambil command dari argument pertama
GOOSE_CMD=${1:-up}

# Validasi: GOOSE_DBSTRING harus ada
if [ -z "$GOOSE_DBSTRING" ]; then
  echo "ERROR: GOOSE_DBSTRING environment variable is not set"
  echo "Example: GOOSE_DBSTRING='postgres://user:pass@host:5432/dbname?sslmode=disable'"
  exit 1
fi

echo "========================================="
echo "Starting database migration"
echo "========================================="
echo "Command: $GOOSE_CMD"
echo "Database Host: $DB_HOST"
echo "Database Port: $DB_PORT"
echo "-----------------------------------------"

# Set goose configuration
export GOOSE_DRIVER=postgres
export GOOSE_MIGRATION_DIR=/app/migrations

# Check database connection
echo "Checking database connection..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U postgres > /dev/null 2>&1; do
    echo "Waiting for database to be ready..."
    sleep 2
done

echo "✓ Database is ready"
echo "-----------------------------------------"
echo "Running goose $GOOSE_CMD..."
echo "GOOSE_DRIVER: $GOOSE_DRIVER"
echo "GOOSE_DBSTRING: ${GOOSE_DBSTRING%%\?*}" # Hide query params for security
echo "GOOSE_MIGRATION_DIR: $GOOSE_MIGRATION_DIR"
echo "-----------------------------------------"

# Run goose command
cd /app
goose "$GOOSE_CMD"

EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "-----------------------------------------"
    echo "✓ Migration completed successfully"
    echo "========================================="
else
    echo "-----------------------------------------"
    echo "✗ Migration failed with exit code: $EXIT_CODE"
    echo "========================================="
    exit $EXIT_CODE
fi