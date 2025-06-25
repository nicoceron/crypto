# Stock Analyzer - Professional Trading Intelligence Platform

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-orange.svg)](https://aws.amazon.com/lambda/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.0+-green.svg)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A modern, serverless stock analysis platform that provides real-time market data, intelligent recommendations, and comprehensive stock ratings analysis. Built with Go microservices architecture and deployed on AWS Lambda for optimal scalability and cost efficiency.

## 🏗️ Architecture Overview

The Stock Analyzer follows a clean architecture pattern with serverless deployment:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Vue.js SPA    │    │   API Gateway    │    │  Lambda Functions│
│                 │────▶│                  │────▶│                 │
│ • Dashboard     │    │ • REST API       │    │ • API Handler   │
│ • Charts        │    │ • CORS           │    │ • Data Ingestion│
│ • Real-time     │    │ • Rate Limiting  │    │ • Scheduler     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
┌─────────────────┐    ┌──────────────────┐             │
│   CloudFront    │    │  CockroachDB     │◀────────────┘
│                 │    │                  │
│ • CDN           │    │ • Serverless     │
│ • HTTPS         │    │ • PostgreSQL     │
│ • Caching       │    │ • Global Scale   │
└─────────────────┘    └──────────────────┘
```

### Core Components

- **Frontend**: Vue.js 3 with TypeScript, responsive design, real-time updates
- **Backend**: Go microservices with clean architecture
- **Database**: CockroachDB Serverless (PostgreSQL-compatible)
- **API**: RESTful API with comprehensive stock market endpoints
- **Infrastructure**: AWS Lambda, API Gateway, CloudFront, S3
- **Data Sources**: Alpaca API (real-time), Custom stock ratings API

## 🚀 Quick Start

### Prerequisites

- **Go** 1.23 or later
- **Node.js** 18 or later
- **Docker** (optional, for local development)
- **AWS CLI** configured with appropriate permissions
- **Terraform** 1.5 or later

### Local Development

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd stock-analyzer/backend
   ```

2. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your API keys and configuration
   ```

3. **Install dependencies**

   ```bash
   go mod download
   ```

4. **Run database migrations**

   ```bash
   export DATABASE_URL="your-cockroachdb-connection-string"
   go run cmd/migrate/main.go
   ```

5. **Start the development server**

   ```bash
   go run cmd/server/main.go
   ```

6. **Run tests**
   ```bash
   make test
   ```

### Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive deployment instructions.

```bash
# Quick deployment
./scripts/setup.sh      # One-time infrastructure setup
./scripts/deploy.sh     # Deploy backend
./scripts/deploy-frontend.sh  # Deploy frontend
```

## 📚 Documentation

### Core Documentation

- [**API Documentation**](docs/API.md) - Complete REST API reference
- [**Architecture Guide**](docs/ARCHITECTURE.md) - Technical architecture details
- [**Database Schema**](docs/DATABASE.md) - Database design and migrations
- [**Configuration Guide**](docs/CONFIGURATION.md) - Environment and configuration
- [**Development Guide**](docs/DEVELOPMENT.md) - Local development setup

### Deployment & Operations

- [**Deployment Guide**](DEPLOYMENT.md) - Production deployment instructions
- [**Infrastructure Guide**](terraform/README.md) - Terraform infrastructure
- [**Monitoring Guide**](docs/MONITORING.md) - Logging and monitoring setup
- [**Troubleshooting**](docs/TROUBLESHOOTING.md) - Common issues and solutions

### Service Documentation

- [**Ingestion Service**](docs/services/INGESTION.md) - Data ingestion pipeline
- [**Recommendation Service**](docs/services/RECOMMENDATION.md) - Stock recommendation engine
- [**Alpaca Integration**](docs/services/ALPACA.md) - Real-time market data

## 🔧 Technology Stack

### Backend

- **Language**: Go 1.23+
- **Framework**: Gin HTTP framework
- **Database**: CockroachDB Serverless (PostgreSQL-compatible)
- **Testing**: testify, go-sqlmock
- **Architecture**: Clean Architecture, Domain-Driven Design

### Frontend

- **Framework**: Vue.js 3 with Composition API
- **Language**: TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Charts**: Chart.js
- **State Management**: Pinia

### Infrastructure

- **Cloud Provider**: AWS
- **Compute**: Lambda Functions
- **API Gateway**: REST API with CORS
- **CDN**: CloudFront
- **Storage**: S3
- **Database**: CockroachDB Serverless
- **IaC**: Terraform

### External APIs

- **Alpaca API**: Real-time stock market data
- **Custom Stock API**: Stock ratings and analyst recommendations
- **Clearbit API**: Company logos and metadata

## 🌟 Key Features

### Real-Time Market Data

- Live stock prices and charts
- Historical price analysis with multiple timeframes
- Market hours detection and indicators

### Intelligent Recommendations

- AI-powered stock recommendations
- Technical analysis signals
- Sentiment scoring and analysis
- Risk assessment and ratings

### Professional Analytics

- Interactive price charts with zoom and pan
- Volume analysis and indicators
- Company fundamentals and ratios
- Analyst ratings aggregation

### Modern User Experience

- Responsive design for desktop and mobile
- Dark/light mode toggle
- Real-time data updates
- Progressive Web App (PWA) capabilities

## 📊 API Endpoints

### Health & Status

- `GET /health` - Application health check
- `GET /api/v1/status` - Detailed system status

### Stock Data

- `GET /api/v1/stocks/{symbol}/price` - Historical price data
- `GET /api/v1/stocks/{symbol}/logo` - Company logo
- `GET /api/v1/stocks/{symbol}/snapshot` - Real-time snapshot

### Ratings & Analysis

- `GET /api/v1/ratings` - Stock ratings with pagination
- `GET /api/v1/ratings/{ticker}` - Ticker-specific ratings
- `GET /api/v1/recommendations` - AI-generated recommendations

### Data Management

- `POST /api/v1/ingest` - Trigger data ingestion
- `GET /api/v1/ingest/status` - Ingestion status

See [API.md](docs/API.md) for complete API documentation with examples.

## 🧪 Testing

The project includes comprehensive testing at multiple levels:

```bash
# Run all tests
make test

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test suites
go test ./internal/api/...          # API tests
go test ./internal/storage/...      # Database tests
go test ./internal/recommendation/... # Business logic tests
```

### Test Coverage Areas

- **Unit Tests**: Service layer, business logic, utilities
- **Integration Tests**: Database operations, external API calls
- **API Tests**: HTTP handlers, middleware, routing
- **Mock Tests**: External dependencies and services

## 🚀 Performance & Scalability

### Cost Optimization

- **Serverless Architecture**: Pay only for actual usage
- **CockroachDB Serverless**: Auto-scaling database with generous free tier
- **CloudFront CDN**: Global content delivery with edge caching
- **Lambda Cold Start Optimization**: Efficient initialization and connection pooling

### Performance Features

- **Response Caching**: Intelligent caching strategies for static and dynamic content
- **Connection Pooling**: Optimized database connections
- **Concurrent Processing**: Goroutine-based concurrent operations
- **Compression**: Gzip compression for API responses

### Monitoring & Observability

- **CloudWatch Integration**: Comprehensive logging and metrics
- **Performance Tracking**: Request duration and throughput monitoring
- **Error Tracking**: Detailed error logging and alerting
- **Health Checks**: Automated health monitoring and alerts

## 🔒 Security

### Authentication & Authorization

- **API Key Management**: Secure API key storage and rotation
- **CORS Configuration**: Proper cross-origin resource sharing setup
- **Rate Limiting**: Request throttling and abuse prevention

### Data Security

- **Encryption at Rest**: Database and S3 encryption
- **Encryption in Transit**: HTTPS/TLS for all communications
- **Secrets Management**: AWS Secrets Manager integration
- **VPC Security**: Network isolation and security groups

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run the test suite
5. Submit a pull request

### Code Style

- Follow Go conventions and best practices
- Use `gofmt` and `golint` for code formatting
- Write comprehensive tests for new features
- Update documentation for API changes

## 📈 Roadmap

### Near-term Features

- [ ] Real-time WebSocket connections
- [ ] Advanced technical indicators
- [ ] Portfolio tracking and management
- [ ] Mobile app development

### Long-term Vision

- [ ] Machine learning-powered predictions
- [ ] Social trading features
- [ ] Multi-broker integration
- [ ] Options and derivatives support

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check our comprehensive [docs](docs/) directory
- **Issues**: Report bugs or request features via GitHub Issues
- **Discussions**: Join our community discussions
- **Email**: Contact the maintainers for enterprise support

---

**Built with ❤️ by the Stock Analyzer Team**

_Last updated: December 2024_
