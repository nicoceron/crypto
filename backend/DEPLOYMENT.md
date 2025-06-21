# Stock Analyzer - Deployment Guide

This document provides comprehensive instructions for deploying the Stock Analyzer application to AWS.

## 🏗️ Architecture Overview

The Stock Analyzer is deployed using a modern serverless architecture:

### Backend Infrastructure

- **AWS Lambda**: 3 serverless functions (API, Ingestion, Scheduler)
- **API Gateway**: RESTful API with CORS support
- **VPC**: Private networking with NAT gateways for security
- **CockroachDB Serverless**: Managed PostgreSQL-compatible database
- **CloudWatch**: Logging and monitoring

### Frontend Infrastructure

- **AWS S3**: Static website hosting for Vue.js application
- **CloudFront**: Global CDN for fast content delivery and HTTPS
- **Automatic deployments**: Build and deploy pipeline

### External Services

- **Alpaca API**: Real-time stock market data
- **Clearbit API**: Company logos and metadata

## 💰 Cost Analysis

**Monthly Costs (USD):**

- AWS Lambda: ~$1-3 (based on usage)
- API Gateway: ~$1-2 (1M requests = $3.50)
- CloudFront: ~$1-3 (1GB transfer = $0.085)
- S3: ~$0.50-1 (storage + requests)
- VPC/NAT Gateway: ~$2-5 (data processing)
- **CockroachDB Serverless**: Free tier (up to 5GB storage)

**Total: ~$5-15/month** (significant savings from previous PostgreSQL setup)

## 📋 Prerequisites

1. **AWS CLI** configured with appropriate permissions
2. **Terraform** >= 1.5
3. **Go** >= 1.21 (for Lambda functions)
4. **Node.js** >= 18 (for frontend)
5. **CockroachDB account** and connection string

## 🚀 Deployment Steps

### 1. Clone and Setup

```bash
git clone <repository-url>
cd stock-analyzer/backend
```

### 2. Configure Environment

Create `terraform/terraform.tfvars`:

```hcl
# Project Configuration
project_name = "stock-analyzer"
environment  = "dev"
aws_region   = "us-west-2"

# Database Configuration (CockroachDB)
cockroachdb_connection_string = "postgresql://username:password@cluster.cockroachlabs.cloud:26257/database?sslmode=verify-full"

# API Keys
alpaca_api_key    = "your-alpaca-api-key"
alpaca_api_secret = "your-alpaca-secret-key"

# Optional: Stock API configuration
stock_api_url   = "https://api.example.com"
stock_api_token = "your-stock-api-token"

# Networking
vpc_cidr = "10.0.0.0/16"
availability_zones = ["us-west-2a", "us-west-2b"]

# Tags
common_tags = {
  Project     = "stock-analyzer"
  Environment = "dev"
  ManagedBy   = "terraform"
  Owner       = "your-name"
}
```

### 3. Deploy Infrastructure

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

This creates:

- VPC with public/private subnets
- Lambda functions with proper IAM roles
- API Gateway with CORS configuration
- S3 bucket for frontend hosting
- CloudFront distribution
- Security groups and networking

### 4. Deploy Backend Code

```bash
cd ..
./scripts/deploy.sh
```

This script:

- Builds Go Lambda functions
- Creates deployment packages
- Updates Lambda function code
- Runs database migrations

### 5. Deploy Frontend

```bash
./scripts/deploy-frontend.sh
```

This script:

- Builds Vue.js production bundle
- Uploads to S3 bucket
- Invalidates CloudFront cache
- Sets proper content types

## 🌐 Application URLs

After deployment, you'll have:

- **Frontend**: `https://your-cloudfront-domain.cloudfront.net`
- **API**: `https://your-api-id.execute-api.region.amazonaws.com/dev`

Get the URLs with:

```bash
cd terraform
terraform output frontend_url
terraform output api_gateway_url
```

## 📊 API Endpoints

### Health Check

```bash
GET /health
```

### Stock Ratings

```bash
GET /api/v1/ratings
GET /api/v1/ratings/{ticker}
```

### Stock Data

```bash
GET /api/v1/stocks/{symbol}/price?period=1W
GET /api/v1/stocks/{symbol}/logo
```

### Recommendations

```bash
GET /api/v1/recommendations
```

### Data Ingestion

```bash
POST /api/v1/ingest
```

## 🔧 Configuration

### Environment Variables

The application uses these environment variables:

**Lambda Functions:**

- `DATABASE_URL`: CockroachDB connection string
- `ALPACA_API_KEY`: Alpaca API key
- `ALPACA_API_SECRET`: Alpaca API secret
- `STOCK_API_URL`: Stock API URL (optional)
- `STOCK_API_TOKEN`: Stock API token (optional)

**Frontend:**

- `VITE_API_BASE_URL`: API Gateway URL

### Database Setup

1. Create CockroachDB Serverless cluster
2. Create database and user
3. Run migrations:

```bash
export DATABASE_URL="your-cockroachdb-connection-string"
go run cmd/migrate/main.go
```

## 🔄 CI/CD Pipeline

### Automated Deployments

The deployment scripts support automated deployments:

```bash
# Deploy everything
./scripts/deploy-all.sh

# Deploy only backend
./scripts/deploy.sh

# Deploy only frontend
./scripts/deploy-frontend.sh
```

### Environment Management

Use different `terraform.tfvars` files for different environments:

```bash
# Development
terraform apply -var-file="dev.tfvars"

# Production
terraform apply -var-file="prod.tfvars"
```

## 📱 Frontend Features

The Vue.js frontend includes:

- **Dashboard**: Market overview and recommendations
- **Stock Details**: Individual stock analysis with charts
- **Real-time Data**: Live price updates via Alpaca API
- **Responsive Design**: Mobile-friendly interface
- **Dark/Light Mode**: User preference support

### Frontend Technology Stack

- **Vue 3**: Progressive JavaScript framework
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and dev server
- **Tailwind CSS**: Utility-first CSS framework
- **Chart.js**: Interactive charts and graphs
- **Pinia**: State management
- **Vue Router**: Client-side routing

## 🛠️ Troubleshooting

### Common Issues

**Lambda Timeout Errors:**

- Check VPC security groups allow outbound traffic on port 26257 (CockroachDB)
- Verify NAT Gateway configuration for internet access

**Frontend Not Loading:**

- Check CloudFront distribution status
- Verify S3 bucket policy allows CloudFront access
- Wait for CloudFront cache invalidation (5-15 minutes)

**API Errors:**

- Verify CockroachDB connection string
- Check Lambda function logs in CloudWatch
- Ensure API Gateway has proper CORS configuration

**Database Connection Issues:**

```bash
# Test connection manually
psql "postgresql://username:password@cluster.cockroachlabs.cloud:26257/database?sslmode=verify-full"
```

### Debugging Commands

```bash
# Check Lambda logs
aws logs tail /aws/lambda/stock-analyzer-dev-api --follow

# Test API endpoints
curl https://your-api-gateway-url/dev/health

# Check CloudFront distribution
aws cloudfront get-distribution --id YOUR_DISTRIBUTION_ID

# Invalidate CloudFront cache
aws cloudfront create-invalidation --distribution-id YOUR_ID --paths "/*"
```

## 🔒 Security Considerations

### Network Security

- Lambda functions run in private subnets
- NAT gateways provide controlled internet access
- Security groups restrict traffic to necessary ports only

### API Security

- CORS properly configured for frontend domain
- API Gateway rate limiting enabled
- CloudWatch logging for audit trails

### Database Security

- CockroachDB uses TLS encryption in transit
- Connection strings stored as Terraform variables
- No direct database access from internet

## 📈 Monitoring and Logging

### CloudWatch Integration

- Lambda function logs automatically captured
- API Gateway access logs enabled
- Custom metrics for application monitoring

### Performance Monitoring

- CloudFront provides CDN analytics
- Lambda execution duration tracking
- Database query performance via CockroachDB console

## 🔄 Updates and Maintenance

### Updating Backend Code

```bash
./scripts/deploy.sh
```

### Updating Frontend

```bash
./scripts/deploy-frontend.sh
```

### Infrastructure Changes

```bash
cd terraform
terraform plan
terraform apply
```

### Database Migrations

```bash
export DATABASE_URL="your-connection-string"
go run cmd/migrate/main.go
```

## 🆘 Support

For issues and questions:

1. Check CloudWatch logs for error details
2. Verify all environment variables are set correctly
3. Ensure all AWS resources are properly configured
4. Test individual components in isolation

## 📚 Additional Resources

- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [CockroachDB Serverless Guide](https://www.cockroachlabs.com/docs/cockroachcloud/serverless)
- [Vue.js Documentation](https://vuejs.org/guide/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
