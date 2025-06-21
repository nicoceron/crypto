#!/bin/bash

# Stock Analyzer Infrastructure Setup Script
# This script helps you set up the infrastructure step by step

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Stock Analyzer Infrastructure Setup${NC}"
echo -e "${BLUE}======================================${NC}\n"

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

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Find project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root if not already there
if [ ! -f "go.mod" ] && [ -f "$PROJECT_ROOT/go.mod" ]; then
    echo -e "${BLUE}📂 Changing to project root directory: $PROJECT_ROOT${NC}"
    cd "$PROJECT_ROOT"
elif [ ! -f "go.mod" ]; then
    print_error "Cannot find project root directory with go.mod file."
    print_error "Please run this script from the project root or scripts directory."
    exit 1
fi

# Step 1: Check prerequisites
echo -e "${BLUE}Step 1: Checking prerequisites...${NC}"

check_command() {
    if command -v $1 &> /dev/null; then
        print_status "$1 is installed"
        return 0
    else
        print_error "$1 is not installed"
        return 1
    fi
}

# Check all required tools
prerequisites_met=true
check_command "go" || prerequisites_met=false
check_command "terraform" || prerequisites_met=false
check_command "aws" || prerequisites_met=false
check_command "jq" || prerequisites_met=false

if [ "$prerequisites_met" = false ]; then
    echo -e "\n${RED}Please install the missing prerequisites and run this script again.${NC}"
    echo -e "\nInstallation guides:"
    echo -e "• Go: https://golang.org/doc/install"
    echo -e "• Terraform: https://learn.hashicorp.com/tutorials/terraform/install-cli"
    echo -e "• AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html"
    echo -e "• jq: https://stedolan.github.io/jq/download/"
    exit 1
fi

# Step 2: Check AWS configuration
echo -e "\n${BLUE}Step 2: Checking AWS configuration...${NC}"

if aws sts get-caller-identity &> /dev/null; then
    AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
    AWS_USER=$(aws sts get-caller-identity --query UserId --output text)
    print_status "AWS CLI is configured for account: $AWS_ACCOUNT (User: $AWS_USER)"
else
    print_error "AWS CLI is not configured or credentials are invalid."
    echo -e "\nPlease run: ${YELLOW}aws configure${NC}"
    exit 1
fi

# Step 3: Set up Go dependencies
echo -e "\n${BLUE}Step 3: Setting up Go dependencies...${NC}"

go mod tidy
print_status "Go dependencies updated"

# Step 4: Set up Terraform variables
echo -e "\n${BLUE}Step 4: Setting up Terraform configuration...${NC}"

if [ ! -f "terraform/terraform.tfvars" ]; then
    cp terraform/terraform.tfvars.example terraform/terraform.tfvars
    print_warning "Created terraform.tfvars from example. Please edit it with your values:"
    echo -e "  ${YELLOW}nano terraform/terraform.tfvars${NC}"
    echo -e "\nRequired values to update:"
    echo -e "  • alpaca_api_key: Your Alpaca API key"
    echo -e "  • alpaca_api_secret: Your Alpaca API secret"
    echo -e "  • stock_api_token: Your stock data API token"
    echo -e "\nPress Enter when you've updated the values..."
    read -r
else
    print_status "terraform.tfvars already exists"
fi

# Step 5: Initialize Terraform
echo -e "\n${BLUE}Step 5: Initializing Terraform...${NC}"

cd terraform
terraform init
print_status "Terraform initialized"

# Step 6: Validate Terraform configuration
echo -e "\n${BLUE}Step 6: Validating Terraform configuration...${NC}"

if terraform validate; then
    print_status "Terraform configuration is valid"
else
    print_error "Terraform configuration validation failed"
    exit 1
fi

# Step 7: Plan infrastructure
echo -e "\n${BLUE}Step 7: Planning infrastructure deployment...${NC}"

if terraform plan -out=tfplan; then
    print_status "Terraform plan completed successfully"
else
    print_error "Terraform planning failed"
    exit 1
fi

# Step 8: Ask for deployment confirmation
echo -e "\n${BLUE}Step 8: Ready to deploy!${NC}"
echo -e "\n${YELLOW}Review the plan above. Do you want to deploy the infrastructure? (y/N)${NC}"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo -e "\n${BLUE}Deploying infrastructure...${NC}"
    
    if terraform apply tfplan; then
        print_status "Infrastructure deployed successfully!"
        
        # Get outputs
        API_URL=$(terraform output -raw api_gateway_url 2>/dev/null || echo "")
        
        echo -e "\n${GREEN}🎉 Deployment completed!${NC}"
        
        if [ -n "$API_URL" ]; then
            echo -e "\n${BLUE}Your API is available at:${NC}"
            echo -e "  $API_URL"
            echo -e "\n${BLUE}Test your deployment:${NC}"
            echo -e "  curl $API_URL/health"
        fi
        
        echo -e "\n${BLUE}Next steps:${NC}"
        echo -e "  1. Build and deploy your Lambda functions:"
        echo -e "     ${YELLOW}./scripts/deploy.sh${NC}"
        echo -e "  2. Run database migrations (see README.md)"
        echo -e "  3. Test your API endpoints"
        
    else
        print_error "Infrastructure deployment failed"
        exit 1
    fi
else
    print_info "Deployment cancelled. You can deploy later with:"
    echo -e "  ${YELLOW}terraform apply tfplan${NC}"
fi

# Cleanup
rm -f tfplan

cd ..

echo -e "\n${GREEN}Setup completed!${NC}"
echo -e "\nFor more information, see the README.md file." 