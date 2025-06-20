# Root Terraform Configuration for Stock Analyzer Application
# This follows clean architecture principles with modular design

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Application = var.app_name
      Environment = var.environment
      ManagedBy   = "terraform"
      Project     = "stock-analyzer"
    }
  }
}

# Generate random suffix for unique resource names
resource "random_id" "suffix" {
  byte_length = 4
}

# Local values for computed names and configurations
locals {
  app_name_full = "${var.app_name}-${var.environment}"
  common_tags = {
    Application = var.app_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "stock-analyzer"
  }
  
  # Database configuration
  db_name     = replace(var.app_name, "-", "_")
  db_username = "stock_admin"
  db_port     = 5432
  
  # Application configuration
  app_port = 8080
  app_name_normalized = lower(replace(var.app_name, "_", "-"))
}

# Infrastructure Layer - VPC and Networking
module "networking" {
  source = "./modules/networking"
  
  app_name         = local.app_name_full
  environment      = var.environment
  vpc_cidr         = var.vpc_cidr
  availability_zones = var.availability_zones
  
  tags = local.common_tags
}

# Data Layer - RDS PostgreSQL Database
module "database" {
  source = "./modules/database"
  
  app_name               = local.app_name_full
  environment           = var.environment
  vpc_id                = module.networking.vpc_id
  private_subnet_ids    = module.networking.private_subnet_ids
  database_subnets      = module.networking.database_subnet_ids
  
  # Database specifications
  db_name               = local.db_name
  db_username           = local.db_username
  db_instance_class     = var.db_instance_class
  db_allocated_storage  = var.db_allocated_storage
  db_max_allocated_storage = var.db_max_allocated_storage
  
  # Security
  allowed_cidr_blocks   = [var.vpc_cidr]
  backup_retention_period = var.db_backup_retention_period
  backup_window         = var.db_backup_window
  maintenance_window    = var.db_maintenance_window
  
  tags = local.common_tags
}

# Application Layer - ECS Fargate Service
module "application" {
  source = "./modules/application"
  
  app_name           = local.app_name_full
  environment        = var.environment
  vpc_id             = module.networking.vpc_id
  public_subnet_ids  = module.networking.public_subnet_ids
  private_subnet_ids = module.networking.private_subnet_ids
  
  # Application configuration
  app_port           = local.app_port
  app_image          = var.app_image
  app_cpu            = var.app_cpu
  app_memory         = var.app_memory
  desired_count      = var.app_desired_count
  min_capacity       = var.app_min_capacity
  max_capacity       = var.app_max_capacity
  
  # Database connection
  database_url       = module.database.database_url
  database_host      = module.database.database_host
  database_port      = local.db_port
  database_name      = local.db_name
  database_username  = local.db_username
  database_password_secret_arn = module.database.database_password_secret_arn
  
  # External API configuration
  stock_api_url      = var.stock_api_url
  stock_api_token    = var.stock_api_token
  alpha_vantage_key  = var.alpha_vantage_key
  alpaca_api_key     = var.alpaca_api_key
  alpaca_api_secret  = var.alpaca_api_secret
  
  tags = local.common_tags
  
  depends_on = [module.database]
}

# Infrastructure Layer - Security (WAF, Security Groups, etc.)
module "security" {
  source = "./modules/security"
  
  app_name             = local.app_name_full
  environment          = var.environment
  alb_arn              = module.application.alb_arn
  allowed_ip_ranges    = var.allowed_ip_ranges
  enable_waf           = var.enable_waf
  
  tags = local.common_tags
}

# Infrastructure Layer - Monitoring and Logging
module "monitoring" {
  source = "./modules/monitoring"
  
  app_name                = local.app_name_full
  environment             = var.environment
  ecs_cluster_name        = module.application.ecs_cluster_name
  ecs_service_name        = module.application.ecs_service_name
  alb_arn_suffix          = module.application.alb_arn_suffix
  target_group_arn_suffix = module.application.target_group_arn_suffix
  
  # Alerting configuration
  notification_email      = var.notification_email
  
  tags = local.common_tags
}

# Utility Layer - Database Migration Job
module "migration" {
  source = "./modules/migration"
  
  app_name                     = local.app_name_full
  environment                  = var.environment
  vpc_id                       = module.networking.vpc_id
  private_subnet_ids           = module.networking.private_subnet_ids
  
  # ECS configuration
  ecs_cluster_arn              = module.application.ecs_cluster_arn
  task_execution_role_arn      = module.application.task_execution_role_arn
  task_role_arn                = module.application.task_role_arn
  
  # Database connection
  database_url                 = module.database.database_url
  database_password_secret_arn = module.database.database_password_secret_arn
  
  # Migration image
  migration_image              = var.migration_image
  
  tags = local.common_tags
  
  depends_on = [module.database, module.application]
} 