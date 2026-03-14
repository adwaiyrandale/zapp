#!/bin/bash
set -e

echo "Starting Payment Gateway development environment..."

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Error: docker-compose is not installed"
    exit 1
fi

# Start infrastructure
echo "Starting infrastructure services..."
if command -v docker-compose &> /dev/null; then
    docker-compose -f configs/docker-compose.yaml up -d
else
    docker compose -f configs/docker-compose.yaml up -d
fi

# Wait for postgres to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Run migrations
echo "Running database migrations..."
go run cmd/migrate/main.go up

# Start services in background
echo "Starting services..."
go run cmd/api/main.go &
go run cmd/payment/main.go &
go run cmd/ledger/main.go &
go run cmd/settlement/main.go &
go run cmd/saga/main.go &

echo ""
echo "Payment Gateway is running!"
echo "API Gateway: http://localhost:8080"
echo "Jaeger UI: http://localhost:16686"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for interrupt
trap "echo 'Stopping services...'; kill 0" INT
wait
