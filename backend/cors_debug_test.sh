#!/bin/bash

echo "🔍 CORS Debug Test - Testing browser console failing endpoints"

# Test the exact endpoints that were failing in the browser console
endpoints=(
  "/api/v1/stocks/NVAX/price?period=1W"
  "/api/v1/stocks/BPMC/price?period=1W" 
  "/api/v1/stocks/ACLX/logo"
  "/api/v1/stocks/TIGO/price?period=1W"
  "/api/v1/recommendations"
  "/api/v1/stocks/SAGE/price?period=1W"
  "/api/v1/stocks/COST/price?period=1W"
)

for endpoint in "${endpoints[@]}"; do
  echo "Testing: $endpoint"
  response=$(curl -s -w "%{http_code}" -H "Origin: https://d2gbyh9p29hcz1.cloudfront.net" \
    "https://idkpnjcw5d.execute-api.us-west-2.amazonaws.com/dev${endpoint}")
  status_code="${response: -3}"
  
  if [ "$status_code" -eq 200 ]; then
    echo "✅ $endpoint - $status_code"
  else
    echo "❌ $endpoint - $status_code"
  fi
done

echo ""
echo "🌐 Browser Cache Clear Instructions:"
echo "1. Open browser dev tools (F12)"
echo "2. Right-click refresh button → 'Empty Cache and Hard Reload'"
echo "3. Or open incognito/private window to test"
