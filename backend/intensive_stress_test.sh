#!/bin/bash

API_URL="https://idkpnjcw5d.execute-api.us-west-2.amazonaws.com/dev"

echo "🚀 Starting INTENSIVE stress test - 50 concurrent requests (25 prices + 25 logos)"

# Top 25 stocks for testing
stocks=(
  "AAPL" "MSFT" "GOOGL" "AMZN" "TSLA" 
  "META" "NVDA" "NFLX" "AAON" "ASML" 
  "ETSY" "DXCM" "RUN" "MLYS" "TTGT" 
  "CGON" "TIGO" "COST" "ACLX" "ADBE"
  "CRM" "PYPL" "INTC" "AMD" "QCOM"
)

start_time=$(date +%s)
success_count=0
error_count=0

# Function to make a request
make_request() {
  local endpoint=$1
  local description=$2  
  local url="${API_URL}${endpoint}"
  
  response=$(curl -s -w "%{http_code}" -H "Origin: https://d2gbyh9p29hcz1.cloudfront.net" "$url")
  status_code="${response: -3}"
  
  if [ "$status_code" -eq 200 ]; then
    echo "✅ $description - $status_code"
    ((success_count++))
  else
    echo "❌ $description - $status_code"
    ((error_count++))
  fi
}

echo "�� Testing 25 stock prices..."
# Test 25 stock prices (with different time periods to vary the load)
periods=("1D" "1W" "1M" "3M" "6M")
for i in "${!stocks[@]}"; do
  stock="${stocks[$i]}"
  period=${periods[$((i % ${#periods[@]}))]}
  make_request "/api/v1/stocks/${stock}/price?period=${period}" "PRICE ${stock} (${period})" &
done

echo "📊 Testing 25 stock logos..."
# Test 25 stock logos
for stock in "${stocks[@]}"; do
  make_request "/api/v1/stocks/${stock}/logo" "LOGO ${stock}" &
done

echo "⏳ Waiting for all 50 requests to complete..."

# Wait for all background jobs to complete
wait

end_time=$(date +%s)
duration=$((end_time - start_time))

echo ""
echo "📊 INTENSIVE Stress Test Results:"
echo "======================================"
echo "Total requests: 50 (25 prices + 25 logos)"
echo "Successful: $success_count"
echo "Failed: $error_count"
echo "Duration: ${duration}s"
echo "Success rate: $(( success_count * 100 / 50 ))%"
echo "Requests/second: $(( 50 / duration ))"

if [ $error_count -gt 5 ]; then
  echo "🚨 HIGH FAILURE RATE - API may be overwhelmed"
  echo "Consider implementing:"
  echo "- Request batching"
  echo "- Client-side caching" 
  echo "- CDN for logos"
elif [ $error_count -gt 0 ]; then
  echo "⚠️  Some requests failed - minor optimization needed"
else
  echo "🎉 PERFECT! All 50 concurrent requests successful!"
  echo "🚀 Your API can handle heavy production load!"
fi

echo ""
echo "💡 This simulates a dashboard with 25 stocks showing prices + logos"
