#!/bin/bash

# Run backend locally with all necessary environment variables
export DATABASE_URL="<YOUR_DATABASE_URL>"
export PORT="8080"
export ALPACA_API_KEY="<YOUR_ALPACA_API_KEY>"
export ALPACA_API_SECRET="<YOUR_ALPACA_SECRET_KEY>"
export STOCK_API_URL="https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list"
export STOCK_API_TOKEN="<YOUR_STOCK_API_TOKEN>"
export ENVIRONMENT="development"
export LOG_LEVEL="info"

echo "🚀 Starting backend server locally..."
echo "📊 Database: CockroachDB Cloud"
echo "🌐 Port: $PORT"
echo "🔍 Health check: http://localhost:$PORT/health"
echo ""

cd ..
go run cmd/lambda/main.go 