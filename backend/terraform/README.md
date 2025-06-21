# Stock Analyzer AWS Infrastructure

This Terraform configuration deploys a Go-based stock analyzer application to AWS using Lambda functions, API Gateway, and RDS PostgreSQL. The infrastructure follows clean architecture principles with modular, reusable components.

## Architecture Overview

The infrastructure consists of the following components:

### 🏗️ Core Modules

1. **Networking Module** (`modules/networking/`)

   - VPC with public and private subnets across multiple AZs
   - NAT Gateways for outbound internet access
   - Security groups for application and database tiers
   - VPC endpoints for cost optimization

2. **Database Module** (`modules/database/`)

   - RDS PostgreSQL instance with encryption
   - Automated backups and monitoring
   - Secrets Manager for credential management
   - Enhanced monitoring and performance insights

3. **Lambda Module** (`modules/lambda/`)

   - Multiple Lambda functions for different responsibilities
   - Scheduled ingestion using EventBridge
   - CloudWatch logging and monitoring
   - VPC integration for database access

4. **API Gateway Module** (`modules/api_gateway/`)
   - REST API Gateway with full CORS support
   - Routes mapping to your Go application endpoints
   - Throttling and caching for production environments
   - CloudWatch logging for monitoring

### 🎯 Lambda Functions

- **API Function**: Handles all HTTP requests from API Gateway
- **Ingestion Function**: Scheduled data ingestion (runs every 4 hours)
- **Scheduler Function**: Handles other scheduled tasks

### 🔒 Security Features

- All resources deployed in private subnets (except NAT gateways)
- Database credentials stored in AWS Secrets Manager
- Security groups with minimal required access
- Encryption at rest for database and S3
- VPC endpoints to reduce NAT gateway costs

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Terraform** >= 1.5 installed
3. **Go** >= 1.23 for building Lambda functions
4. **API Keys** for external services (Alpaca, stock data API)

## Quick Start

### 1. Clone and Navigate

```bash
cd terraform
```

### 2. Configure Variables

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your actual values
```

### 3. Initialize Terraform

```bash
terraform init
```

### 4. Plan Deployment

```bash
terraform plan
```

### 5. Deploy Infrastructure

```bash
terraform apply
```

### 6. Build and Deploy Lambda Functions

```bash
# Build your Go application for Lambda
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go
zip lambda-deployment.zip bootstrap

# Upload to the created S3 bucket
aws s3 cp lambda-deployment.zip s3://$(terraform output -raw s3_bucket_name)/

# Update Lambda functions
aws lambda update-function-code \
  --function-name $(terraform output -json lambda_functions | jq -r '.api.function_name') \
  --s3-bucket $(terraform output -raw s3_bucket_name) \
  --s3-key lambda-deployment.zip
```

## Configuration

### Required Variables

Create a `terraform.tfvars` file with these required variables:

```hcl
# Sensitive - Get from Alpaca dashboard
alpaca_api_key    = "your-alpaca-api-key"
alpaca_api_secret = "your-alpaca-api-secret"

# Sensitive - Get from your stock data provider
stock_api_token   = "your-stock-api-token"
```

### Optional Variables

```hcl
# AWS Configuration
aws_region = "us-west-2"  # Change to your preferred region

# Environment Configuration
environment = "dev"  # dev, staging, prod

# Database Configuration - CockroachDB Serverless
cockroachdb_connection_string = "postgresql://username:password@cluster.cockroachlabs.cloud:26257/defaultdb?sslmode=require"

# Network Configuration
vpc_cidr = "10.0.0.0/16"
availability_zones = ["us-west-2a", "us-west-2b"]
```

## Deployment Environments

### Development

```bash
terraform workspace new dev
terraform apply -var="environment=dev"
```

### Production

```bash
terraform workspace new prod
terraform apply -var="environment=prod"
```

## Lambda Function Development

### Structure for Lambda

Your Go application needs to be adapted for Lambda. Create a main function that handles Lambda events:

```go
// cmd/lambda/main.go
package main

import (
    "context"
    "encoding/json"
    "os"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/gin-gonic/gin"
    ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"

    "stock-analyzer/internal/api"
    // ... other imports
)

var ginLambda *ginadapter.GinLambda

func init() {
    // Set Gin to release mode in Lambda
    gin.SetMode(gin.ReleaseMode)

    // Initialize your router
    router := api.SetupRouter(
        // ... your dependencies
    )

    ginLambda = ginadapter.New(router)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
    lambda.Start(Handler)
}
```

### Building for Lambda

```bash
# Build for Linux (Lambda runtime)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap cmd/lambda/main.go

# Create deployment package
zip lambda-deployment.zip bootstrap

# Upload to S3
aws s3 cp lambda-deployment.zip s3://your-deployment-bucket/
```

## Database Migration

After deployment, run your database migrations:

```bash
# Run migrations using your CockroachDB connection string
export DATABASE_URL="your-cockroachdb-connection-string"
go run cmd/migrate/main.go
```

## Monitoring and Logging

### CloudWatch Logs

- Lambda functions: `/aws/lambda/stock-analyzer-{env}-{function}`
- API Gateway: `/aws/apigateway/stock-analyzer-{env}`
- CockroachDB: Built-in monitoring via CockroachDB Cloud Console

### CloudWatch Metrics

- Lambda: Duration, errors, throttles
- API Gateway: Request count, latency, errors
- CockroachDB: Built-in metrics via CockroachDB Cloud Console

### Alarms (Production)

```bash
# Set up CloudWatch alarms for production monitoring
aws cloudwatch put-metric-alarm \
  --alarm-name "stock-analyzer-api-errors" \
  --alarm-description "API Gateway 5xx errors" \
  --metric-name 5XXError \
  --namespace AWS/ApiGateway \
  --statistic Sum \
  --period 300 \
  --threshold 5 \
  --comparison-operator GreaterThanThreshold
```

## Cost Optimization

### Development Environment

- Uses CockroachDB Serverless free tier
- Minimal Lambda memory allocation
- No API Gateway caching
- Short log retention (7 days)

### Production Optimizations

- CockroachDB Serverless auto-scaling
- API Gateway caching enabled
- VPC endpoints to reduce NAT costs
- Longer log retention for compliance

### Estimated Monthly Costs (us-west-2)

- **Development**: ~$5-15/month (CockroachDB free tier)
- **Production**: ~$20-50/month (depends on traffic)

## Troubleshooting

### Common Issues

1. **Lambda timeout connecting to CockroachDB**

   - Check security group rules (port 26257)
   - Verify Lambda is in correct subnets with NAT Gateway
   - Check VPC configuration

2. **API Gateway 502 errors**

   - Check Lambda function logs
   - Verify Lambda response format
   - Check function timeout settings

3. **CockroachDB connection refused**
   - Verify CockroachDB cluster is running
   - Check connection string format
   - Verify SSL configuration

### Debugging Commands

```bash
# Check Lambda logs
aws logs tail /aws/lambda/stock-analyzer-dev-api --follow

# Test API Gateway
curl -X GET "$(terraform output -raw api_gateway_url)/health"

# Test CockroachDB connection
export DATABASE_URL="your-cockroachdb-connection-string"
go run cmd/migrate/main.go
```

## Security Best Practices

1. **Never commit sensitive values** to version control
2. **Use AWS Secrets Manager** for database credentials
3. **Enable CloudTrail** for audit logging
4. **Regularly rotate** API keys and passwords
5. **Use least privilege** IAM policies
6. **Enable VPC Flow Logs** for network monitoring

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

⚠️ **Warning**: This will permanently delete all data, including the database!

## Support

For issues related to:

- **Infrastructure**: Check CloudWatch logs and AWS documentation
- **Application**: Review your Go application logs
- **Database**: Check RDS logs and connection parameters

## Contributing

When adding new resources:

1. Follow the existing module structure
2. Add appropriate tags and naming conventions
3. Include monitoring and logging
4. Update this README with new configuration options
