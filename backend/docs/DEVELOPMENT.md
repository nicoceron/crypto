# Development Guide

This guide provides comprehensive information for developers working on the Stock Analyzer project, including local setup, development workflow, testing strategies, and contribution guidelines.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Development Environment Setup](#development-environment-setup)
3. [Project Structure](#project-structure)
4. [Development Workflow](#development-workflow)
5. [Testing Strategy](#testing-strategy)
6. [Code Style Guidelines](#code-style-guidelines)
7. [Debugging](#debugging)
8. [Contributing](#contributing)
9. [Release Process](#release-process)

## Quick Start

### Prerequisites

Ensure you have the following installed:

```bash
# Check Go version
go version  # Should be 1.23 or later

# Check Node.js version (for frontend)
node --version  # Should be 18 or later

# Check Docker (optional)
docker --version

# Check AWS CLI (for deployment)
aws --version

# Check Terraform (for infrastructure)
terraform --version
```

### 5-Minute Setup

```bash
# 1. Clone and setup
git clone <repository-url>
cd stock-analyzer/backend

# 2. Install dependencies
go mod download

# 3. Copy environment template
cp .env.example .env
# Edit .env with your API keys

# 4. Run database migrations
export DATABASE_URL="your-cockroachdb-connection-string"
go run cmd/migrate/main.go

# 5. Start development server
go run cmd/server/main.go

# 6. Run tests
make test
```

## Development Environment Setup

### 1. Go Environment

```bash
# Install Go 1.23+
# macOS
brew install go

# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Verify installation
go version
go env GOPATH
```

### 2. Environment Variables

Create `.env` file in project root:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:26257/stock_data_dev?sslmode=require

# External APIs
ALPACA_API_KEY=PKTEST_your_sandbox_key
ALPACA_API_SECRET=your_sandbox_secret
STOCK_API_TOKEN=your_test_token

# Application
ENVIRONMENT=development
LOG_LEVEL=debug
PORT=8080
```

### 3. Database Setup

#### Option A: CockroachDB Cloud (Recommended)

```bash
# 1. Sign up for CockroachDB Cloud free tier
# 2. Create a serverless cluster
# 3. Download the connection string
# 4. Update DATABASE_URL in .env
```

#### Option B: Local CockroachDB

```bash
# Install CockroachDB locally
curl https://binaries.cockroachdb.com/cockroach-latest.linux-amd64.tgz | tar -xz
sudo cp -i cockroach-latest.linux-amd64/cockroach /usr/local/bin/

# Start local cluster
cockroach start-single-node --insecure --listen-addr=localhost:26257 --http-addr=localhost:8080

# Create database
cockroach sql --insecure -e "CREATE DATABASE stock_data_dev;"

# Update DATABASE_URL
DATABASE_URL=postgresql://root@localhost:26257/stock_data_dev?sslmode=disable
```

### 4. IDE Setup

#### VS Code Configuration

Create `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.useLanguageServer": true,
  "go.lintOnSave": "package",
  "go.vetOnSave": "package",
  "editor.formatOnSave": true,
  "go.buildOnSave": "package",
  "go.testFlags": ["-v", "-race"]
}
```

Create `.vscode/extensions.json`:

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-vscode.makefile-tools"
  ]
}
```

#### GoLand Configuration

1. Import project from existing sources
2. Enable Go modules support
3. Configure code style:
   - Use goimports for imports
   - Enable golangci-lint
   - Set line length to 100

## Project Structure

```
stock-analyzer/backend/
├── cmd/                          # Application entry points
│   ├── lambda/main.go           # Lambda function handler
│   ├── migrate/main.go          # Database migrations
│   └── server/main.go           # HTTP server
├── internal/                     # Private application code
│   ├── api/                     # HTTP handlers and routing
│   │   ├── handlers.go
│   │   ├── handlers_test.go
│   │   ├── middleware.go
│   │   └── router.go
│   ├── domain/                  # Domain models and interfaces
│   │   ├── interfaces.go
│   │   └── models.go
│   ├── storage/                 # Database implementations
│   │   ├── postgres.go
│   │   └── postgres_test.go
│   ├── alpaca/                  # External API adapters
│   │   ├── service.go
│   │   └── service_test.go
│   ├── ingestion/               # Data ingestion service
│   │   ├── service.go
│   │   └── service_test.go
│   └── recommendation/          # Recommendation engine
│       ├── service.go
│       └── service_test.go
├── pkg/                         # Public packages
│   ├── config/                  # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   └── errors/                  # Error handling
│       └── errors.go
├── docs/                        # Documentation
├── scripts/                     # Build and deployment scripts
├── terraform/                   # Infrastructure as code
├── migrations/                  # Database migrations
├── .env.example                 # Environment template
├── go.mod                       # Go modules
├── go.sum                       # Dependency checksums
├── Makefile                     # Build targets
└── README.md                    # Project overview
```

### Architecture Layers

#### 1. Domain Layer (`internal/domain/`)

- **Pure business logic**
- **No external dependencies**
- Contains entities, value objects, and interfaces

#### 2. Application Layer (`internal/api/`)

- **Use cases and orchestration**
- HTTP handlers and middleware
- Request/response mapping

#### 3. Infrastructure Layer (`internal/storage/`, `internal/alpaca/`)

- **External integrations**
- Database repositories
- External API adapters

## Development Workflow

### 1. Feature Development

```bash
# 1. Create feature branch
git checkout -b feature/add-portfolio-tracking

# 2. Make changes with tests
# - Write failing test
# - Implement feature
# - Make test pass
# - Refactor

# 3. Run tests locally
make test

# 4. Run linting
golangci-lint run

# 5. Commit changes
git add .
git commit -m "feat: add portfolio tracking functionality"

# 6. Push and create PR
git push origin feature/add-portfolio-tracking
```

### 2. Code Review Process

1. **Self Review**

   - Run all tests
   - Check code coverage
   - Review your own changes

2. **Automated Checks**

   - GitHub Actions CI/CD
   - Unit tests
   - Integration tests
   - Code quality checks

3. **Peer Review**
   - At least one approval required
   - Address feedback
   - Ensure documentation is updated

### 3. Testing Workflow

```bash
# Run all tests
make test

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/api/...

# Run tests in watch mode (with entr)
find . -name "*.go" | entr -r go test ./...

# Benchmark tests
go test -bench=. ./internal/recommendation/
```

### 4. Database Development

```bash
# Create new migration
# migrations/004_new_feature.sql

# Run migrations
go run cmd/migrate/main.go

# Rollback (if supported)
# Implement down migrations manually

# Test with fresh database
dropdb stock_data_test && createdb stock_data_test
DATABASE_URL="postgresql://user@localhost/stock_data_test" go run cmd/migrate/main.go
```

## Testing Strategy

### Test Types

#### 1. Unit Tests

- **Target**: Individual functions and methods
- **Location**: `*_test.go` files alongside source
- **Scope**: Fast, isolated tests

```go
func TestStockRepository_GetStockRatings(t *testing.T) {
    // Arrange
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    repo := NewPostgresRepository(db)

    // Setup mock expectations
    rows := sqlmock.NewRows([]string{"rating_id", "ticker", "company"}).
        AddRow("123", "AAPL", "Apple Inc.")
    mock.ExpectQuery("SELECT (.+) FROM stock_ratings").
        WillReturnRows(rows)

    // Act
    ratings, err := repo.GetStockRatings(context.Background(), FilterOptions{})

    // Assert
    assert.NoError(t, err)
    assert.Len(t, ratings, 1)
    assert.Equal(t, "AAPL", ratings[0].Ticker)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

#### 2. Integration Tests

- **Target**: Component interactions
- **Location**: `*_integration_test.go`
- **Scope**: Database, external APIs

```go
//go:build integration
// +build integration

func TestStockRepository_Integration(t *testing.T) {
    // Requires actual database connection
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := NewPostgresRepository(db)

    // Test actual database operations
    rating := domain.StockRating{
        Ticker:  "AAPL",
        Company: "Apple Inc.",
        // ... other fields
    }

    err := repo.CreateStockRating(context.Background(), rating)
    assert.NoError(t, err)

    ratings, err := repo.GetStockRatings(context.Background(), FilterOptions{
        Ticker: "AAPL",
    })
    assert.NoError(t, err)
    assert.Len(t, ratings, 1)
}
```

#### 3. API Tests

- **Target**: HTTP endpoints
- **Location**: `internal/api/handlers_test.go`
- **Scope**: Request/response flow

```go
func TestHandlers_GetStockPrice(t *testing.T) {
    // Setup
    mockRepo := &MockStockRepository{}
    mockAlpaca := &MockAlpacaService{}
    handlers := NewHandlers(mockRepo, nil, nil, mockAlpaca)

    router := gin.New()
    router.GET("/api/v1/stocks/:symbol/price", handlers.GetStockPrice)

    // Test
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/v1/stocks/AAPL/price?period=1W", nil)
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)

    var response StockPriceResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "AAPL", response.Symbol)
}
```

### Test Organization

```bash
# Run unit tests only
go test -short ./...

# Run integration tests
go test -tags=integration ./...

# Run with coverage
go test -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Mocking

#### Using go-sqlmock for Database

```go
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    return db, mock
}
```

#### Creating Custom Mocks

```go
type MockAlpacaService struct {
    mock.Mock
}

func (m *MockAlpacaService) GetHistoricalBars(ctx context.Context, symbol string, timeframe string, start, end time.Time) ([]domain.PriceBar, error) {
    args := m.Called(ctx, symbol, timeframe, start, end)
    return args.Get(0).([]domain.PriceBar), args.Error(1)
}
```

## Code Style Guidelines

### Go Conventions

1. **Naming**

   ```go
   // Good
   type StockRepository interface {}
   func GetStockRatings() {}
   var stockData []Stock

   // Bad
   type stockrepository interface {}
   func getstock_ratings() {}
   var stock_data []Stock
   ```

2. **Package Names**

   ```go
   // Good
   package domain
   package storage

   // Bad
   package domainmodels
   package db_storage
   ```

3. **Error Handling**

   ```go
   // Good
   if err != nil {
       return nil, fmt.Errorf("failed to get stock ratings: %w", err)
   }

   // Bad
   if err != nil {
       panic(err)
   }
   ```

4. **Comments**
   ```go
   // GetStockRatings retrieves stock ratings with optional filtering
   // and pagination. Returns ErrNotFound if no ratings are found.
   func GetStockRatings(ctx context.Context, filters FilterOptions) ([]StockRating, error) {
       // Implementation
   }
   ```

### Project Standards

#### 1. Function Length

- Keep functions under 50 lines
- Extract helper functions when needed
- One responsibility per function

#### 2. File Organization

- Group related functionality
- Keep test files alongside source
- Use consistent naming patterns

#### 3. Dependencies

- Prefer standard library
- Keep external dependencies minimal
- Document dependency decisions

### Code Quality Tools

#### 1. golangci-lint Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - golint
    - gosec
    - ineffassign
    - misspell
    - unparam
    - unused
    - staticcheck

linters-settings:
  golint:
    min-confidence: 0.8
  goimports:
    local-prefixes: stock-analyzer
```

#### 2. Pre-commit Hooks

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go-fmt
        entry: gofmt -l -s -w
        language: system
        files: \.go$
      - id: go-lint
        name: go-lint
        entry: golangci-lint run
        language: system
        files: \.go$
```

## Debugging

### 1. Local Debugging

#### Using Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug main application
dlv debug cmd/server/main.go

# Debug tests
dlv test ./internal/api/
```

#### VS Code Debugging

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/server/main.go",
      "env": {
        "DATABASE_URL": "postgresql://localhost:26257/stock_data_dev"
      }
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/api"
    }
  ]
}
```

### 2. Logging

#### Structured Logging

```go
import (
    "github.com/sirupsen/logrus"
)

func (h *Handlers) GetStockPrice(c *gin.Context) {
    symbol := c.Param("symbol")

    logger := logrus.WithFields(logrus.Fields{
        "symbol":    symbol,
        "handler":   "GetStockPrice",
        "requestID": c.GetHeader("X-Request-ID"),
    })

    logger.Info("processing stock price request")

    // ... implementation

    logger.WithField("barsCount", len(bars)).Info("returning price data")
}
```

#### Debug Logging

```go
// Enable debug logging in development
logger.SetLevel(logrus.DebugLevel)

logger.Debug("fetching data from Alpaca API")
logger.Debugf("query parameters: %+v", params)
```

### 3. Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/recommendation/

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/recommendation/

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Contributing

### 1. Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Add tests
6. Update documentation
7. Submit a pull request

### 2. Pull Request Guidelines

#### PR Title Format

```
type(scope): description

Examples:
feat(api): add portfolio tracking endpoints
fix(database): resolve connection pool exhaustion
docs(readme): update installation instructions
```

#### PR Description Template

```markdown
## Summary

Brief description of the changes

## Changes

- List of specific changes
- Bullet points for each major change

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Documentation

- [ ] API documentation updated
- [ ] README updated if needed
- [ ] Architecture docs updated if needed

## Breaking Changes

List any breaking changes and migration steps
```

### 3. Code Review Checklist

#### For Authors

- [ ] All tests pass
- [ ] Code coverage maintained/improved
- [ ] Documentation updated
- [ ] No hardcoded secrets
- [ ] Error handling implemented
- [ ] Logging added where appropriate

#### For Reviewers

- [ ] Code follows project conventions
- [ ] Tests are comprehensive
- [ ] Documentation is clear
- [ ] Performance impact considered
- [ ] Security implications reviewed

## Release Process

### 1. Version Management

```bash
# Create release branch
git checkout -b release/v1.2.0

# Update version in relevant files
# Update CHANGELOG.md

# Tag release
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0
```

### 2. Deployment Pipeline

1. **Development**

   - Feature branches
   - Automated testing
   - Code review

2. **Staging**

   - Release candidate testing
   - Integration testing
   - Performance testing

3. **Production**
   - Tagged releases
   - Rollback capability
   - Monitoring alerts

### 3. Hotfix Process

```bash
# Create hotfix branch from main
git checkout -b hotfix/v1.2.1 main

# Make critical fix
# Add tests
# Update version

# Deploy to production
# Merge back to main and develop
```

---

_Development Guide v1.0 - Last updated: December 2024_
