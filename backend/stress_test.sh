#!/bin/bash

API_URL="https://idkpnjcw5d.execute-api.us-west-2.amazonaws.com/dev"

echo "🧪 Starting stress test - 20 concurrent requests (simulating one page load)"

# Simulate the requests that happen on the main page
endpoints=(
  "/api/v1/ratings?limit=20&page=1&sort_by=updated_at&order=desc"
  "/api/v1/recommendations"
  "/api/v1/stocks/AAPL/logo"
  "/api/v1/stocks/MSFT/logo"
  "/api/v1/stocks/GOOGL/logo"
  "/api/v1/stocks/AMZN/logo"
  "/api/v1/stocks/TSLA/logo"
  "/api/v1/stocks/META/logo"
  "/api/v1/stocks/NVDA/logo"
  "/api/v1/stocks/NFLX/logo"
  "/api/v1/stocks/AAON/logo"
  "/api/v1/stocks/ASML/logo"
  "/api/v1/stocks/ETSY/logo"
  "/api/v1/stocks/DXCM/logo"
  "/api/v1/stocks/RUN/logo"
  "/api/v1/stocks/MLYS/logo"
  "/api/v1/stocks/TTGT/logo"
  "/api/v1/stocks/CGON/logo"
  "/api/v1/stocks/TIGO/logo"
  "/api/v1/stocks/COST/logo"
  "/api/v1/stocks/ACLX/logo"
)

start_time=$(date +%s)
success_count=0
error_count=0

# Function to make a request
make_request() {
  local endpoint=$1
  local url="${API_URL}${endpoint}"
  
  response=$(curl -s -w "%{http_code}" -H "Origin: https://d2gbyh9p29hcz1.cloudfront.net" "$url")
  status_code="${response: -3}"
  
  if [ "$status_code" -eq 200 ]; then
    echo "✅ $endpoint - $status_code"
    ((success_count++))
  else
    echo "❌ $endpoint - $status_code"
    ((error_count++))
  fi
}

# Run all requests in parallel
for endpoint in "${endpoints[@]}"; do
  make_request "$endpoint" &
done

# Wait for all background jobs to complete
wait

end_time=$(date +%s)
duration=$((end_time - start_time))

echo ""
echo "📊 Stress Test Results:"
echo "Total requests: ${#endpoints[@]}"
echo "Successful: $success_count"
echo "Failed: $error_count"
echo "Duration: ${duration}s"
echo "Success rate: $(( success_count * 100 / ${#endpoints[@]} ))%"

if [ $error_count -gt 0 ]; then
  echo "⚠️  Some requests failed - may need rate limiting adjustments"
else
  echo "🎉 All requests successful!"
fi
