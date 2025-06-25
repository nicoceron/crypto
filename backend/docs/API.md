# API Documentation

Stock Analyzer provides a RESTful API for accessing stock market data, ratings, and recommendations. The API is designed following OpenAPI 3.0 specifications and provides consistent responses with proper error handling.

## Base URL

- **Production**: `https://your-api-gateway-url.execute-api.region.amazonaws.com/dev`
- **Development**: `http://localhost:8080`

## Authentication

Currently, the API is publicly accessible without authentication. Future versions will include API key authentication.

## Request/Response Format

All API endpoints accept and return JSON-formatted data. Requests should include the appropriate `Content-Type` header:

```http
Content-Type: application/json
```

## Error Handling

The API uses standard HTTP status codes and returns detailed error information:

```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": "symbol parameter is required",
  "timestamp": "2024-12-24T12:00:00Z"
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Endpoints

### Health Check

#### GET /health

Check the health status of the API.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2024-12-24T12:00:00Z",
  "version": "1.0.0"
}
```

---

### Stock Price Data

#### GET /api/v1/stocks/{symbol}/price

Retrieve historical price data for a specific stock symbol.

**Parameters:**

- `symbol` (path, required): Stock symbol (e.g., AAPL, MSFT)
- `period` (query, optional): Time period for data
  - `1W` - 1 week (hourly data)
  - `1M` - 1 month (hourly data)
  - `3M` - 3 months (daily data)
  - `6M` - 6 months (daily data)
  - `1Y` - 1 year (daily data)
  - `2Y` - 2 years (daily data)
  - Default: `1M`

**Example Request:**

```bash
curl -X GET "https://api.example.com/api/v1/stocks/AAPL/price?period=1W" \
  -H "Content-Type: application/json"
```

**Example Response:**

```json
{
  "symbol": "AAPL",
  "bars": [
    {
      "timestamp": "2024-12-20T09:30:00Z",
      "open": 150.25,
      "high": 152.8,
      "low": 149.9,
      "close": 151.45,
      "volume": 2500000
    },
    {
      "timestamp": "2024-12-20T10:30:00Z",
      "open": 151.45,
      "high": 153.2,
      "low": 151.0,
      "close": 152.75,
      "volume": 1800000
    }
  ]
}
```

---

### Stock Logo

#### GET /api/v1/stocks/{symbol}/logo

Retrieve the company logo URL for a stock symbol.

**Parameters:**

- `symbol` (path, required): Stock symbol (e.g., AAPL, MSFT)

**Example Request:**

```bash
curl -X GET "https://api.example.com/api/v1/stocks/AAPL/logo"
```

**Example Response:**

```json
{
  "symbol": "AAPL",
  "logo_url": "https://logo.clearbit.com/apple.com"
}
```

**Response Headers:**

- `Cache-Control: public, max-age=3600`
- `ETag: "AAPL"`

---

### Stock Ratings

#### GET /api/v1/ratings

Retrieve paginated stock ratings with optional filtering and sorting.

**Parameters:**

- `page` (query, optional): Page number (default: 1)
- `limit` (query, optional): Items per page (default: 20, max: 100)
- `sort_by` (query, optional): Sort field
  - `time` - Sort by rating time
  - `ticker` - Sort by ticker symbol
  - `updated_at` - Sort by last update
  - Default: `time`
- `order` (query, optional): Sort order (`asc` or `desc`, default: `desc`)
- `ticker` (query, optional): Filter by ticker symbol
- `action` (query, optional): Filter by action type
  - `upgrade` - Rating upgrades
  - `downgrade` - Rating downgrades
  - `initiate` - New coverage
  - `maintain` - Maintained ratings

**Example Request:**

```bash
curl -X GET "https://api.example.com/api/v1/ratings?page=1&limit=10&sort_by=time&order=desc&action=upgrade"
```

**Example Response:**

```json
{
  "data": [
    {
      "rating_id": "123e4567-e89b-12d3-a456-426614174000",
      "ticker": "AAPL",
      "company": "Apple Inc.",
      "brokerage": "Goldman Sachs",
      "action": "upgrade",
      "rating_from": "Hold",
      "rating_to": "Buy",
      "target_from": 150.0,
      "target_to": 180.0,
      "time": "2024-12-24T08:30:00Z",
      "created_at": "2024-12-24T08:35:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total_items": 1250,
    "total_pages": 125
  }
}
```

---

#### GET /api/v1/ratings/{ticker}

Retrieve all ratings for a specific stock ticker.

**Parameters:**

- `ticker` (path, required): Stock ticker symbol
- `page` (query, optional): Page number (default: 1)
- `limit` (query, optional): Items per page (default: 20)

**Example Request:**

```bash
curl -X GET "https://api.example.com/api/v1/ratings/AAPL?page=1&limit=5"
```

**Example Response:**

```json
{
  "data": [
    {
      "rating_id": "123e4567-e89b-12d3-a456-426614174000",
      "ticker": "AAPL",
      "company": "Apple Inc.",
      "brokerage": "Goldman Sachs",
      "action": "upgrade",
      "rating_from": "Hold",
      "rating_to": "Buy",
      "target_from": 150.0,
      "target_to": 180.0,
      "time": "2024-12-24T08:30:00Z",
      "created_at": "2024-12-24T08:35:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 5,
    "total_items": 45,
    "total_pages": 9
  }
}
```

---

### Stock Recommendations

#### GET /api/v1/recommendations

Retrieve AI-generated stock recommendations based on analysis of ratings, price movements, and market data.

**Parameters:**

- `limit` (query, optional): Number of recommendations to return (default: 10, max: 50)

**Example Request:**

```bash
curl -X GET "https://api.example.com/api/v1/recommendations?limit=5"
```

**Example Response:**

```json
{
  "recommendations": [
    {
      "ticker": "AAPL",
      "company": "Apple Inc.",
      "score": 8.5,
      "rationale": "Strong analyst upgrades with positive momentum and solid fundamentals",
      "latest_rating": "Buy",
      "target_price": 180.0,
      "technical_signal": "bullish",
      "sentiment_score": 0.75,
      "generated_at": "2024-12-24T12:00:00Z"
    },
    {
      "ticker": "MSFT",
      "company": "Microsoft Corporation",
      "score": 8.2,
      "rationale": "Consistent growth in cloud services with strong earnings outlook",
      "latest_rating": "Outperform",
      "target_price": 420.0,
      "technical_signal": "bullish",
      "sentiment_score": 0.72,
      "generated_at": "2024-12-24T12:00:00Z"
    }
  ],
  "generated_at": "2024-12-24T12:00:00Z",
  "total_analyzed": 1250,
  "data_freshness": "5 minutes ago"
}
```

---

### Data Ingestion

#### POST /api/v1/ingest

Trigger manual data ingestion from external sources.

**Request Body:** None required

**Example Request:**

```bash
curl -X POST "https://api.example.com/api/v1/ingest" \
  -H "Content-Type: application/json"
```

**Example Response:**

```json
{
  "message": "Data ingestion started successfully",
  "job_id": "ingest-123456789",
  "estimated_duration": "2-5 minutes",
  "status": "started",
  "timestamp": "2024-12-24T12:00:00Z"
}
```

---

## Rate Limiting

The API implements rate limiting to ensure fair usage and system stability:

- **Default Limit**: 100 requests per minute per IP
- **Burst Limit**: 20 requests per 10 seconds

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## CORS Policy

The API supports Cross-Origin Resource Sharing (CORS) with the following configuration:

- **Allowed Origins**: `*` (all origins)
- **Allowed Methods**: `GET, POST, PUT, DELETE, OPTIONS`
- **Allowed Headers**: `Content-Type, Authorization, X-Requested-With`
- **Max Age**: 86400 seconds (24 hours)

## Caching

The API implements intelligent caching to improve performance:

### Cache Headers

- **Stock Logos**: `Cache-Control: public, max-age=3600` (1 hour)
- **Price Data**: `Cache-Control: public, max-age=300` (5 minutes)
- **Ratings**: `Cache-Control: public, max-age=600` (10 minutes)

### ETags

ETags are provided for cacheable resources to enable conditional requests:

```http
ETag: "AAPL-logo-v1"
```

Client can use `If-None-Match` header to check for changes:

```http
If-None-Match: "AAPL-logo-v1"
```

## SDK and Libraries

### JavaScript/TypeScript

```typescript
import { StockAnalyzerAPI } from "@stock-analyzer/api-client";

const api = new StockAnalyzerAPI("https://api.example.com");

// Get stock price
const priceData = await api.getStockPrice("AAPL", "1M");

// Get recommendations
const recommendations = await api.getRecommendations(10);
```

### Python

```python
from stock_analyzer import StockAnalyzerAPI

api = StockAnalyzerAPI('https://api.example.com')

# Get stock ratings
ratings = api.get_ratings(page=1, limit=20, ticker='AAPL')

# Get logo
logo = api.get_stock_logo('AAPL')
```

## Webhooks

Future versions will support webhooks for real-time notifications:

- **New Ratings**: Triggered when new analyst ratings are added
- **Price Alerts**: Triggered when stock prices cross specified thresholds
- **Recommendation Updates**: Triggered when AI recommendations change

## OpenAPI Specification

The complete API specification is available in OpenAPI 3.0 format:

- **Swagger UI**: `https://api.example.com/docs`
- **OpenAPI JSON**: `https://api.example.com/openapi.json`

## Support

For API support and questions:

- **Documentation**: [https://docs.stock-analyzer.com](https://docs.stock-analyzer.com)
- **GitHub Issues**: [https://github.com/your-repo/issues](https://github.com/your-repo/issues)
- **Email**: api-support@stock-analyzer.com

---

_API Documentation v1.0 - Last updated: December 2024_
