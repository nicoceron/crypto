#!/bin/bash

# Stock Analyzer Lambda Deployment Script
# This script builds and deploys the Go Lambda functions to AWS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="stock-analyzer"
ENVIRONMENT=${ENVIRONMENT:-dev}
AWS_REGION=${AWS_REGION:-us-west-2}

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$ROOT_DIR/build"
TERRAFORM_DIR="$ROOT_DIR/terraform"

echo -e "${BLUE}🚀 Starting deployment for $PROJECT_NAME ($ENVIRONMENT)${NC}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}🔍 Checking prerequisites...${NC}"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.23 or later."
        exit 1
    fi
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install AWS CLI."
        exit 1
    fi
    
    # Check if terraform is installed
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed. Please install Terraform."
        exit 1
    fi
    
    # Find and change to project root directory
if [ ! -f "$ROOT_DIR/go.mod" ] && [ -f "$(dirname "$ROOT_DIR")/go.mod" ]; then
    echo -e "${BLUE}📂 Adjusting to project root directory${NC}"
    ROOT_DIR="$(dirname "$ROOT_DIR")"
    BUILD_DIR="$ROOT_DIR/build"
    TERRAFORM_DIR="$ROOT_DIR/terraform"
elif [ ! -f "$ROOT_DIR/go.mod" ]; then
    print_error "Cannot find project root directory with go.mod file."
    print_error "Please run this script from the project root or scripts directory."
    exit 1
fi
    
    print_status "All prerequisites satisfied"
}

# Build Lambda functions
build_lambda() {
    echo -e "${BLUE}🔨 Building Lambda function...${NC}"
    
    # Create build directory
    mkdir -p "$BUILD_DIR"
    
    # Clean previous builds
    rm -f "$BUILD_DIR/bootstrap" "$BUILD_DIR/lambda-deployment.zip"
    
    # Build for Linux (Lambda runtime)
    cd "$ROOT_DIR"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BUILD_DIR/bootstrap" cmd/lambda/main.go
    
    # Create deployment package
    cd "$BUILD_DIR"
    zip -r lambda-deployment.zip bootstrap
    
    print_status "Lambda function built successfully"
}

# Get Terraform outputs
get_terraform_outputs() {
    echo -e "${BLUE}📋 Getting Terraform outputs...${NC}"
    
    cd "$TERRAFORM_DIR"
    
    # Check if Terraform has been initialized
    if [ ! -d ".terraform" ]; then
        print_error "Terraform not initialized. Please run 'terraform init' first."
        exit 1
    fi
    
    # Get outputs
    S3_BUCKET=$(terraform output -raw s3_bucket_name 2>/dev/null || echo "")
    API_FUNCTION_NAME=$(terraform output -json lambda_functions 2>/dev/null | jq -r '.api.function_name' || echo "")
    INGESTION_FUNCTION_NAME=$(terraform output -json lambda_functions 2>/dev/null | jq -r '.ingestion.function_name' || echo "")
    SCHEDULER_FUNCTION_NAME=$(terraform output -json lambda_functions 2>/dev/null | jq -r '.scheduler.function_name' || echo "")
    API_GATEWAY_URL=$(terraform output -raw api_gateway_url 2>/dev/null || echo "")
    
    if [ -z "$S3_BUCKET" ] || [ -z "$API_FUNCTION_NAME" ]; then
        print_error "Failed to get Terraform outputs. Make sure infrastructure is deployed."
        exit 1
    fi
    
    print_status "Terraform outputs retrieved"
}

# Upload to S3
upload_to_s3() {
    echo -e "${BLUE}☁️  Uploading to S3...${NC}"
    
    # Upload deployment package
    aws s3 cp "$BUILD_DIR/lambda-deployment.zip" "s3://$S3_BUCKET/lambda-deployment.zip" --region "$AWS_REGION"
    
    print_status "Deployment package uploaded to S3"
}

# Update Lambda functions
update_lambda_functions() {
    echo -e "${BLUE}🔄 Updating Lambda functions...${NC}"
    
    # Update API function
    aws lambda update-function-code \
        --function-name "$API_FUNCTION_NAME" \
        --s3-bucket "$S3_BUCKET" \
        --s3-key "lambda-deployment.zip" \
        --region "$AWS_REGION" \
        --no-cli-pager > /dev/null
    
    print_status "API function updated"
    
    # Update ingestion function
    aws lambda update-function-code \
        --function-name "$INGESTION_FUNCTION_NAME" \
        --s3-bucket "$S3_BUCKET" \
        --s3-key "lambda-deployment.zip" \
        --region "$AWS_REGION" \
        --no-cli-pager > /dev/null
    
    print_status "Ingestion function updated"
    
    # Update scheduler function
    aws lambda update-function-code \
        --function-name "$SCHEDULER_FUNCTION_NAME" \
        --s3-bucket "$S3_BUCKET" \
        --s3-key "lambda-deployment.zip" \
        --region "$AWS_REGION" \
        --no-cli-pager > /dev/null
    
    print_status "Scheduler function updated"
}

# Wait for functions to be ready
wait_for_functions() {
    echo -e "${BLUE}⏳ Waiting for functions to be ready...${NC}"
    
    for func in "$API_FUNCTION_NAME" "$INGESTION_FUNCTION_NAME" "$SCHEDULER_FUNCTION_NAME"; do
        aws lambda wait function-updated --function-name "$func" --region "$AWS_REGION"
    done
    
    print_status "All functions are ready"
}

# Test deployment
test_deployment() {
    echo -e "${BLUE}🧪 Testing deployment...${NC}"
    
    if [ -n "$API_GATEWAY_URL" ]; then
        # Test health endpoint
        response=$(curl -s -o /dev/null -w "%{http_code}" "$API_GATEWAY_URL/health" || echo "000")
        
        if [ "$response" = "200" ]; then
            print_status "Health check passed"
        else
            print_warning "Health check failed (HTTP $response)"
        fi
    else
        print_warning "API Gateway URL not found, skipping health check"
    fi
}

# Run database migrations
run_migrations() {
    echo -e "${BLUE}🗃️  Running database migrations...${NC}"
    
    cd "$TERRAFORM_DIR"
    DB_ENDPOINT=$(terraform output -raw database_endpoint 2>/dev/null || echo "")
    
    if [ -n "$DB_ENDPOINT" ]; then
        print_warning "Database endpoint found: $DB_ENDPOINT"
        print_warning "Please run migrations manually using your preferred migration tool"
        print_warning "Example: migrate -database 'postgres://user:pass@$DB_ENDPOINT:5432/stock_data?sslmode=require' -path ./migrations up"
    else
        print_warning "Database endpoint not found in Terraform outputs"
    fi
}

# Print deployment summary
print_summary() {
    echo -e "\n${GREEN}🎉 Deployment completed successfully!${NC}\n"
    
    if [ -n "$API_GATEWAY_URL" ]; then
        echo -e "${BLUE}API Gateway URL:${NC} $API_GATEWAY_URL"
        echo -e "${BLUE}Health Check:${NC} $API_GATEWAY_URL/health"
        echo -e "${BLUE}Example API Call:${NC} curl '$API_GATEWAY_URL/api/v1/ratings'"
    fi
    
    echo -e "\n${BLUE}Lambda Functions:${NC}"
    echo -e "  • API: $API_FUNCTION_NAME"
    echo -e "  • Ingestion: $INGESTION_FUNCTION_NAME"
    echo -e "  • Scheduler: $SCHEDULER_FUNCTION_NAME"
    
    echo -e "\n${BLUE}Monitoring:${NC}"
    echo -e "  • CloudWatch Logs: /aws/lambda/$PROJECT_NAME-$ENVIRONMENT-*"
    echo -e "  • API Gateway Logs: /aws/apigateway/$PROJECT_NAME-$ENVIRONMENT"
    
    echo -e "\n${YELLOW}Next Steps:${NC}"
    echo -e "  1. Run database migrations if needed"
    echo -e "  2. Monitor CloudWatch logs for any issues"
    echo -e "  3. Test your API endpoints"
}

# Main deployment flow
main() {
    check_prerequisites
    build_lambda
    get_terraform_outputs
    upload_to_s3
    update_lambda_functions
    wait_for_functions
    test_deployment
    run_migrations
    print_summary
}

# Handle script arguments
case "${1:-}" in
    "build")
        build_lambda
        ;;
    "deploy")
        get_terraform_outputs
        upload_to_s3
        update_lambda_functions
        wait_for_functions
        test_deployment
        ;;
    "test")
        get_terraform_outputs
        test_deployment
        ;;
    *)
        main
        ;;
esac 