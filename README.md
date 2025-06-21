# Stock Analyzer

A full-stack stock analysis and recommendation platform that provides real-time market data, analyst ratings, and AI-generated investment recommendations.

<img width="1393" alt="image" src="https://github.com/user-attachments/assets/746b4032-9354-46be-b4d2-9d1eb7903417" />


## 🚀 Features

### 📊 Market Analysis

- **Real-time Stock Data**: Live price charts and historical data via Alpaca API
- **Analyst Ratings**: Comprehensive database of stock ratings from major brokerages
- **Price Targets**: Target price tracking and analysis
- **Technical Indicators**: Advanced chart analysis with multiple timeframes

### 🤖 AI Recommendations

- **Smart Recommendations**: AI-powered investment suggestions based on multiple data sources
- **Sentiment Analysis**: News sentiment scoring and market mood tracking
- **Risk Assessment**: Comprehensive risk analysis and scoring
- **Portfolio Insights**: Personalized investment recommendations

### 💻 Modern Interface

- **Responsive Design**: Beautiful, mobile-first interface built with Vue 3 and Tailwind CSS
- **Interactive Charts**: Dynamic price charts with Chart.js integration
- **Real-time Updates**: Live data streaming and automatic refresh
- **Advanced Filtering**: Powerful search and filtering capabilities

## 🏗️ Architecture

### Backend (Go)

- **Framework**: Gin HTTP router with clean architecture
- **Database**: PostgreSQL with CockroachDB Cloud support
- **APIs**: Alpaca Markets integration for real-time data
- **Services**: Modular service architecture with domain-driven design

### Frontend (Vue.js)

- **Framework**: Vue 3 with TypeScript and Composition API
- **Styling**: Tailwind CSS with Headless UI components
- **State Management**: Pinia for reactive state management
- **Routing**: Vue Router with lazy loading
- **Testing**: Vitest + Cypress for unit and E2E testing

## 🛠️ Development Setup

### Prerequisites

- **Go 1.23+**
- **Node.js 18+**
- **CockroachDB**
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
   ALPACA_BASE_URL="https://paper-api.alpaca.markets"  # or live URL

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
   go run cmd/server/main.go
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

- `GET /api/stocks/price/:symbol` - Historical price data
- `GET /api/stocks/logo/:symbol` - Company logo
- `GET /api/stocks/ratings` - Paginated stock ratings
- `GET /api/stocks/ratings/:ticker` - Ratings for specific stock

### Recommendations

- `GET /api/recommendations` - AI-generated recommendations
- `POST /api/ingestion/trigger` - Trigger data ingestion

### System

- `GET /api/health` - Health check endpoint

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

### Current Deployment Plan

#### Frontend Deployment (Static Hosting)

- [ ] Build production bundle: `npm run build`
- [ ] Deploy `dist/` folder to AWS S3 bucket
- [ ] Configure S3 for static website hosting
- [ ] Set up CloudFront CDN for global distribution

#### Backend Deployment (Containerized)

- [ ] **GitHub Actions Pipeline**: Automated CI/CD
  - Build Docker image on push to main
  - Push image to Amazon ECR
  - Deploy to AWS App Runner
- [ ] **Future Migration**: Consider AWS Lambda for serverless architecture
- [ ] **Database**: CockroachDB Cloud (already configured)

### Quick Deployment Commands

#### Build Docker Image

```bash
cd backend
docker build -t stock-analyzer-backend .
```

#### Frontend Production Build

```bash
cd frontend
npm run build
# Upload dist/ folder to S3
```

## 🔧 Configuration

### Environment Variables

#### Backend (.env)

```env
DATABASE_URL=            # PostgreSQL connection string
ALPACA_API_KEY=         # Alpaca Markets API key
ALPACA_SECRET_KEY=      # Alpaca Markets secret key
ALPACA_BASE_URL=        # Alpaca API base URL
PORT=8080               # Server port
GIN_MODE=release        # Gin mode (debug/release)
```

#### Frontend (.env)

```env
VITE_API_BASE_URL=      # Backend API URL
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and clean architecture principles
- Use TypeScript for all frontend code
- Write comprehensive tests for new features
- Update documentation for API changes
- Follow conventional commit messages

## 📋 Roadmap

### Current Sprint

- [x] Core stock data ingestion
- [x] Real-time price charts
- [x] Basic recommendation engine
- [x] Responsive frontend interface

### Next Release

- [ ] Enhanced AI recommendation algorithms
- [ ] Portfolio tracking and management
- [ ] Real-time notifications and alerts
- [ ] Advanced technical analysis indicators
- [ ] Social sentiment integration

### Future Features

- [ ] Mobile app (React Native)
- [ ] Options trading analysis
- [ ] Cryptocurrency support
- [ ] Paper trading simulation
- [ ] Community features and social trading

## 📚 Tech Stack

### Backend

- **Language**: Go 1.23+
- **Framework**: Gin HTTP router
- **Database**: PostgreSQL / CockroachDB Cloud
- **External APIs**: Alpaca Markets
- **Architecture**: Clean Architecture / Domain-Driven Design

### Frontend

- **Framework**: Vue 3 with TypeScript
- **Styling**: Tailwind CSS + Headless UI
- **Charts**: Chart.js + Vue-ChartJS
- **State**: Pinia
- **Testing**: Vitest + Cypress
- **Build**: Vite

### DevOps

- **Containerization**: Docker
- **CI/CD**: GitHub Actions
- **Cloud**: AWS (S3, ECR, App Runner)
- **Database**: CockroachDB Cloud
- **Monitoring**: Built-in health checks

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support, please open an issue on GitHub or contact the development team.

---

**Built with ❤️ for the trading community**
