# Stock Analyzer

A stock analysis platform that provides stock ratings data, price charts, and basic recommendations based on analyst consensus.

<img width="1393" alt="image" src="https://github.com/user-attachments/assets/746b4032-9354-46be-b4d2-9d1eb7903417" />

## 🚀 Current Features

### 📊 Stock Data

- **Stock Ratings**: Database of analyst ratings from major brokerages with search and pagination
- **Price Charts**: Interactive price charts with multiple timeframes (1W, 1M, 3M, 6M, 1Y, 2Y) via Alpaca API
- **Company Logos**: Automatic logo fetching using Clearbit API
- **Rating History**: Track analyst rating changes and price targets over time

### 🤖 Basic Recommendations

- **Analyst-Based Recommendations**: Simple recommendation engine based on recent analyst ratings
- **Rating Scoring**: Weighted scoring system based on rating quality and recency
- **Trending Stocks**: Display of recently upgraded stocks and top recommendations

### 💻 Modern Interface

- **Responsive Design**: Mobile-first interface built with Vue 3 and Tailwind CSS
- **Interactive Charts**: Price charts with Chart.js integration
- **Real-time Data**: Live price data from Alpaca Markets API
- **Search & Filtering**: Search by ticker or company name with pagination

## 🏗️ Architecture

### Backend (Go)

- **Framework**: Gin HTTP router with clean architecture
- **Database**: PostgreSQL/CockroachDB with migrations
- **APIs**: Alpaca Markets for price data, external API for ratings
- **Services**: Modular service architecture with ingestion, recommendation, and storage layers
- **Deployment**: AWS Lambda with API Gateway (fully configured)

### Frontend (Vue.js)

- **Framework**: Vue 3 with TypeScript and Composition API
- **Styling**: Tailwind CSS with Headless UI components
- **State Management**: Pinia for reactive state management
- **Routing**: Vue Router with views for Markets, Recommendations, and Stock Details
- **Charts**: Chart.js with vue-chartjs for price visualization

### Database Schema

- **stock_ratings**: Core table with analyst ratings, price targets, and metadata
- **enriched_stock_data**: Additional data storage for future enhancements
- **Indexes**: Optimized for search and filtering performance

## 🛠️ Development Setup

### Prerequisites

- **Go 1.23+**
- **Node.js 18+**
- **PostgreSQL** or **CockroachDB**
- **Alpaca Markets API Key**

### Backend Setup

1. **Clone and navigate to backend**:

   ```bash
   cd backend
   ```

2. **Install dependencies**:

   ```bash
   go mod tidy
   ```

3. **Environment Configuration**:
   Create `.env` file:

   ```env
   # Database
   DATABASE_URL="postgresql://username:password@localhost:5432/stock_data?sslmode=disable"

   # Alpaca API
   ALPACA_API_KEY="your_alpaca_api_key"
   ALPACA_SECRET_KEY="your_alpaca_secret_key"

   # Stock Ratings API
   STOCK_API_URL="your_stock_api_url"
   STOCK_API_TOKEN="your_stock_api_token"

   # Server
   PORT="8080"
   GIN_MODE="debug"
   ```

4. **Database Migration**:

   ```bash
   go run cmd/migrate/main.go
   ```

5. **Run the server**:
   ```bash
   go run cmd/lambda/main.go
   ```

The backend will be available at `http://localhost:8080`

### Frontend Setup

1. **Navigate to frontend**:

   ```bash
   cd frontend
   ```

2. **Install dependencies**:

   ```bash
   npm install
   ```

3. **Environment Configuration**:
   Create `.env` file:

   ```env
   VITE_API_BASE_URL=http://localhost:8080
   ```

4. **Development server**:

   ```bash
   npm run dev
   ```

5. **Build for production**:
   ```bash
   npm run build
   ```

The frontend will be available at `http://localhost:5173`

## 🔌 API Endpoints

### Stock Data

- `GET /api/v1/stocks/:symbol/price` - Historical price data with period parameter
- `GET /api/v1/stocks/:symbol/logo` - Company logo URL
- `GET /api/v1/ratings` - Paginated stock ratings with search
- `GET /api/v1/ratings/:ticker` - Ratings for specific stock

### Recommendations

- `GET /api/v1/recommendations` - Basic analyst-based recommendations

### Admin

- `POST /api/v1/ingest` - Trigger data ingestion
- `GET /health` - Health check endpoint

## 🧪 Testing

### Backend Tests

```bash
cd backend
go test ./...
```

### Frontend Tests

```bash
cd frontend

# Unit tests
npm run test:unit

# E2E tests
npm run test:e2e

# Linting
npm run lint
```

## 📦 Deployment

### AWS Lambda Deployment (Configured)

The application is configured for AWS Lambda deployment with:

- **Lambda Functions**: API, Ingestion, and Scheduler functions
- **API Gateway**: HTTP API with CORS support
- **S3**: Deployment package storage
- **EventBridge**: Scheduled ingestion triggers
- **Terraform**: Complete infrastructure as code

#### Deploy to AWS

```bash
cd backend/terraform
terraform init
terraform plan
terraform apply

# Deploy Lambda functions
cd ../scripts
./deploy.sh
```

#### Frontend Deployment

Build and deploy the frontend to any static hosting service:

```bash
cd frontend
npm run build
# Upload dist/ folder to your hosting service
```

## 🔧 Configuration

### Environment Variables

#### Backend (.env)

```env
DATABASE_URL=            # PostgreSQL connection string
ALPACA_API_KEY=         # Alpaca Markets API key
ALPACA_SECRET_KEY=      # Alpaca Markets secret key
STOCK_API_URL=          # Stock ratings API URL
STOCK_API_TOKEN=        # Stock ratings API token
PORT=8080               # Server port (local development)
GIN_MODE=release        # Gin mode (debug/release)
```

#### Frontend (.env)

```env
VITE_API_BASE_URL=      # Backend API URL
```

## 📚 Tech Stack

### Backend

- **Language**: Go 1.23+
- **Framework**: Gin HTTP router
- **Database**: PostgreSQL / CockroachDB
- **External APIs**: Alpaca Markets
- **Architecture**: Clean Architecture / Domain-Driven Design
- **Deployment**: AWS Lambda + API Gateway

### Frontend

- **Framework**: Vue 3 with TypeScript
- **Styling**: Tailwind CSS + Headless UI
- **Charts**: Chart.js + vue-chartjs
- **State**: Pinia
- **Testing**: Vitest + Cypress
- **Build**: Vite

### Infrastructure

- **Cloud**: AWS (Lambda, API Gateway, S3, EventBridge)
- **IaC**: Terraform
- **Database**: CockroachDB Cloud
- **CI/CD**: Deployment scripts

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**A modern stock analysis platform built with Go and Vue.js**
