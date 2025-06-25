terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = var.common_tags
  }
}

# Data sources for existing resources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Networking module - VPC, subnets, security groups
module "networking" {
  source = "./modules/networking"
  
  project_name        = var.project_name
  environment         = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  common_tags        = var.common_tags
}

# Database configuration - Using CockroachDB Serverless
# No infrastructure needed - CockroachDB is managed externally

# Lambda module - Lambda functions and related resources
module "lambda" {
  source = "./modules/lambda"
  
  project_name           = var.project_name
  environment            = var.environment
  vpc_id                = module.networking.vpc_id
  app_subnet_ids        = module.networking.public_subnet_ids
  app_security_group_id = module.networking.app_security_group_id
  database_url          = var.cockroachdb_connection_string
  common_tags           = var.common_tags
  
  # Application configuration
  alpaca_api_key    = var.alpaca_api_key
  alpaca_api_secret = var.alpaca_api_secret
  stock_api_url     = var.stock_api_url
  stock_api_token   = var.stock_api_token
}

# API Gateway module
module "api_gateway" {
  source = "./modules/api_gateway"
  
  project_name    = var.project_name
  environment     = var.environment
  lambda_functions = module.lambda.lambda_functions
  common_tags     = var.common_tags
}

# Frontend module for S3 + CloudFront hosting
module "frontend" {
  source = "./modules/frontend"
  
  project_name     = var.project_name
  environment      = var.environment
  api_gateway_url  = module.api_gateway.api_gateway_url
  
  common_tags = var.common_tags
}

# S3 bucket for Lambda deployment packages
resource "aws_s3_bucket" "lambda_deployments" {
  bucket = "${var.project_name}-${var.environment}-lambda-deployments-${random_string.bucket_suffix.result}"
  
  tags = var.common_tags
}

resource "aws_s3_bucket_versioning" "lambda_deployments" {
  bucket = aws_s3_bucket.lambda_deployments.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "lambda_deployments" {
  bucket = aws_s3_bucket.lambda_deployments.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "lambda_deployments" {
  bucket = aws_s3_bucket.lambda_deployments.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "random_string" "bucket_suffix" {
  length  = 8
  special = false
  upper   = false
} 