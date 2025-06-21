#!/bin/bash

# Deploy Frontend to AWS S3 + CloudFront
# This script builds the Vue.js frontend and deploys it to AWS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/../frontend"
TERRAFORM_DIR="$PROJECT_ROOT/terraform"

echo -e "${BLUE}🚀 Starting Frontend Deployment${NC}"
echo "Project Root: $PROJECT_ROOT"
echo "Frontend Dir: $FRONTEND_DIR"

# Check if frontend directory exists
if [ ! -d "$FRONTEND_DIR" ]; then
    echo -e "${RED}❌ Frontend directory not found: $FRONTEND_DIR${NC}"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "$TERRAFORM_DIR/terraform.tfstate" ]; then
    echo -e "${RED}❌ Terraform state not found. Please run terraform apply first.${NC}"
    exit 1
fi

# Get Terraform outputs
echo -e "${YELLOW}📋 Getting deployment information from Terraform...${NC}"
cd "$TERRAFORM_DIR"

# Check if frontend module is deployed
if ! terraform output frontend_s3_bucket >/dev/null 2>&1; then
    echo -e "${RED}❌ Frontend infrastructure not found. Please run terraform apply first.${NC}"
    exit 1
fi

S3_BUCKET=$(terraform output -raw frontend_s3_bucket)
CLOUDFRONT_ID=$(terraform output -raw cloudfront_distribution_id)
API_URL=$(terraform output -raw api_gateway_url)
FRONTEND_URL=$(terraform output -raw frontend_url)

echo "S3 Bucket: $S3_BUCKET"
echo "CloudFront ID: $CLOUDFRONT_ID"
echo "API URL: $API_URL"
echo "Frontend URL: $FRONTEND_URL"

# Build the frontend
echo -e "${YELLOW}🔨 Building Vue.js frontend...${NC}"
cd "$FRONTEND_DIR"

# Update environment variables - use direct API Gateway (CloudFront routing has persistent issues)
echo "VITE_API_BASE_URL=$API_URL" > .env.production

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Installing dependencies...${NC}"
    npm install
fi

# Build the project
echo -e "${YELLOW}🏗️ Building production bundle...${NC}"
npm run build

# Check if build was successful
if [ ! -d "dist" ]; then
    echo -e "${RED}❌ Build failed - dist directory not found${NC}"
    exit 1
fi

# Deploy to S3
echo -e "${YELLOW}☁️ Uploading to S3...${NC}"
aws s3 sync dist/ s3://$S3_BUCKET --delete --exact-timestamps

# Set correct content types for specific files
echo -e "${YELLOW}🔧 Setting content types...${NC}"
aws s3 cp s3://$S3_BUCKET/index.html s3://$S3_BUCKET/index.html --content-type "text/html" --cache-control "no-cache" --metadata-directive REPLACE
aws s3 cp s3://$S3_BUCKET/ s3://$S3_BUCKET/ --recursive --exclude "*" --include "*.js" --content-type "application/javascript" --cache-control "max-age=31536000" --metadata-directive REPLACE
aws s3 cp s3://$S3_BUCKET/ s3://$S3_BUCKET/ --recursive --exclude "*" --include "*.css" --content-type "text/css" --cache-control "max-age=31536000" --metadata-directive REPLACE
aws s3 cp s3://$S3_BUCKET/ s3://$S3_BUCKET/ --recursive --exclude "*" --include "*.png" --content-type "image/png" --cache-control "max-age=31536000" --metadata-directive REPLACE
aws s3 cp s3://$S3_BUCKET/ s3://$S3_BUCKET/ --recursive --exclude "*" --include "*.jpg" --content-type "image/jpeg" --cache-control "max-age=31536000" --metadata-directive REPLACE
aws s3 cp s3://$S3_BUCKET/ s3://$S3_BUCKET/ --recursive --exclude "*" --include "*.svg" --content-type "image/svg+xml" --cache-control "max-age=31536000" --metadata-directive REPLACE

# Invalidate CloudFront cache
echo -e "${YELLOW}🔄 Invalidating CloudFront cache...${NC}"
INVALIDATION_ID=$(aws cloudfront create-invalidation --distribution-id $CLOUDFRONT_ID --paths "/*" --query 'Invalidation.Id' --output text)
echo "Invalidation ID: $INVALIDATION_ID"

echo -e "${GREEN}✅ Frontend deployment complete!${NC}"
echo ""
echo -e "${BLUE}📊 Deployment Summary:${NC}"
echo -e "Frontend URL: ${GREEN}$FRONTEND_URL${NC}"
echo -e "API URL: ${GREEN}$API_URL${NC}"
echo -e "S3 Bucket: ${GREEN}$S3_BUCKET${NC}"
echo -e "CloudFront Distribution: ${GREEN}$CLOUDFRONT_ID${NC}"
echo ""
echo -e "${YELLOW}⏳ Note: CloudFront invalidation may take 5-15 minutes to complete.${NC}"
echo -e "${BLUE}🌐 Your frontend will be available at: $FRONTEND_URL${NC}" 