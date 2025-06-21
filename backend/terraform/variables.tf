variable "aws_region" {
  description = "AWS region where resources will be created"
  type        = string
  default     = "us-west-2"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "stock-analyzer"
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b"]
}

variable "cockroachdb_connection_string" {
  description = "CockroachDB connection string"
  type        = string
  sensitive   = true
}

# Application-specific variables
variable "alpaca_api_key" {
  description = "Alpaca API key"
  type        = string
  sensitive   = true
}

variable "alpaca_api_secret" {
  description = "Alpaca API secret"
  type        = string
  sensitive   = true
}

variable "stock_api_url" {
  description = "Stock API URL"
  type        = string
  default     = "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list"
}

variable "stock_api_token" {
  description = "Stock API token"
  type        = string
  sensitive   = true
}

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default = {
    Project     = "stock-analyzer"
    Environment = "dev"
    ManagedBy   = "terraform"
  }
} 