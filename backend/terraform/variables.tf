# Variables for Stock Analyzer Infrastructure
# Organized by concern following clean architecture principles

# ============================================================================
# CORE APPLICATION VARIABLES
# ============================================================================

variable "app_name" {
  description = "Name of the application"
  type        = string
  default     = "stock-analyzer"
  
  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.app_name))
    error_message = "App name must contain only lowercase letters, numbers, and hyphens."
  }
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

# ============================================================================
# NETWORKING CONFIGURATION
# ============================================================================

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
  
  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "VPC CIDR must be a valid IPv4 CIDR block."
  }
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b", "us-west-2c"]
  
  validation {
    condition     = length(var.availability_zones) >= 2
    error_message = "At least 2 availability zones must be specified for high availability."
  }
}

# ============================================================================
# DATABASE CONFIGURATION
# ============================================================================

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
  
  validation {
    condition = contains([
      "db.t3.micro", "db.t3.small", "db.t3.medium", "db.t3.large",
      "db.t3.xlarge", "db.t3.2xlarge", "db.r5.large", "db.r5.xlarge",
      "db.r5.2xlarge", "db.r5.4xlarge"
    ], var.db_instance_class)
    error_message = "DB instance class must be a valid RDS instance type."
  }
}

variable "db_allocated_storage" {
  description = "Initial database storage in GB"
  type        = number
  default     = 20
  
  validation {
    condition     = var.db_allocated_storage >= 20 && var.db_allocated_storage <= 65536
    error_message = "Database storage must be between 20 and 65536 GB."
  }
}

variable "db_max_allocated_storage" {
  description = "Maximum database storage for autoscaling in GB"
  type        = number
  default     = 100
  
  validation {
    condition     = var.db_max_allocated_storage >= var.db_allocated_storage
    error_message = "Max allocated storage must be greater than or equal to allocated storage."
  }
}

variable "db_backup_retention_period" {
  description = "Database backup retention period in days"
  type        = number
  default     = 7
  
  validation {
    condition     = var.db_backup_retention_period >= 0 && var.db_backup_retention_period <= 35
    error_message = "Backup retention period must be between 0 and 35 days."
  }
}

variable "db_backup_window" {
  description = "Database backup window (UTC)"
  type        = string
  default     = "03:00-04:00"
  
  validation {
    condition     = can(regex("^([0-1][0-9]|2[0-3]):[0-5][0-9]-([0-1][0-9]|2[0-3]):[0-5][0-9]$", var.db_backup_window))
    error_message = "Backup window must be in format HH:MM-HH:MM."
  }
}

variable "db_maintenance_window" {
  description = "Database maintenance window (UTC)"
  type        = string
  default     = "sun:04:00-sun:05:00"
  
  validation {
    condition = can(regex("^(sun|mon|tue|wed|thu|fri|sat):[0-2][0-9]:[0-5][0-9]-(sun|mon|tue|wed|thu|fri|sat):[0-2][0-9]:[0-5][0-9]$", var.db_maintenance_window))
    error_message = "Maintenance window must be in format ddd:HH:MM-ddd:HH:MM."
  }
}

# ============================================================================
# APPLICATION CONFIGURATION
# ============================================================================

variable "app_image" {
  description = "Docker image for the application"
  type        = string
  default     = "public.ecr.aws/docker/library/golang:1.23-alpine"
  # In production, this would be your custom ECR image
}

variable "migration_image" {
  description = "Docker image for database migrations"
  type        = string
  default     = "public.ecr.aws/docker/library/golang:1.23-alpine"
  # In production, this would be your custom migration image
}

variable "app_cpu" {
  description = "CPU units for the application (1024 = 1 vCPU)"
  type        = number
  default     = 512
  
  validation {
    condition = contains([
      256, 512, 1024, 2048, 4096, 8192, 16384
    ], var.app_cpu)
    error_message = "CPU must be one of the valid Fargate CPU values."
  }
}

variable "app_memory" {
  description = "Memory in MB for the application"
  type        = number
  default     = 1024
  
  validation {
    condition = (
      (var.app_cpu == 256 && contains([512, 1024, 2048], var.app_memory)) ||
      (var.app_cpu == 512 && var.app_memory >= 1024 && var.app_memory <= 4096) ||
      (var.app_cpu == 1024 && var.app_memory >= 2048 && var.app_memory <= 8192) ||
      (var.app_cpu == 2048 && var.app_memory >= 4096 && var.app_memory <= 16384) ||
      (var.app_cpu == 4096 && var.app_memory >= 8192 && var.app_memory <= 30720) ||
      (var.app_cpu >= 8192 && var.app_memory >= 16384 && var.app_memory <= 61440)
    )
    error_message = "Memory must be compatible with the specified CPU value according to Fargate specifications."
  }
}

variable "app_desired_count" {
  description = "Desired number of application instances"
  type        = number
  default     = 2
  
  validation {
    condition     = var.app_desired_count >= 1 && var.app_desired_count <= 100
    error_message = "Desired count must be between 1 and 100."
  }
}

variable "app_min_capacity" {
  description = "Minimum number of application instances for auto scaling"
  type        = number
  default     = 1
  
  validation {
    condition     = var.app_min_capacity >= 1 && var.app_min_capacity <= var.app_desired_count
    error_message = "Min capacity must be at least 1 and not exceed desired count."
  }
}

variable "app_max_capacity" {
  description = "Maximum number of application instances for auto scaling"
  type        = number
  default     = 10
  
  validation {
    condition     = var.app_max_capacity >= var.app_desired_count && var.app_max_capacity <= 100
    error_message = "Max capacity must be at least the desired count and not exceed 100."
  }
}

# ============================================================================
# EXTERNAL API CONFIGURATION
# ============================================================================

variable "stock_api_url" {
  description = "URL for the external stock ratings API"
  type        = string
  default     = "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list"
}

variable "stock_api_token" {
  description = "JWT token for the external stock ratings API"
  type        = string
  sensitive   = true
  default     = ""
}

variable "alpha_vantage_key" {
  description = "Alpha Vantage API key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "alpaca_api_key" {
  description = "Alpaca API key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "alpaca_api_secret" {
  description = "Alpaca API secret"
  type        = string
  sensitive   = true
  default     = ""
}

# ============================================================================
# SECURITY CONFIGURATION
# ============================================================================

variable "allowed_ip_ranges" {
  description = "List of IP ranges allowed to access the application"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # In production, restrict this to specific IPs
  
  validation {
    condition = alltrue([
      for ip in var.allowed_ip_ranges : can(cidrhost(ip, 0))
    ])
    error_message = "All IP ranges must be valid CIDR blocks."
  }
}

variable "enable_waf" {
  description = "Enable AWS WAF for additional security"
  type        = bool
  default     = true
}

# ============================================================================
# MONITORING CONFIGURATION
# ============================================================================

variable "notification_email" {
  description = "Email address for CloudWatch alerts"
  type        = string
  default     = ""
  
  validation {
    condition = var.notification_email == "" || can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.notification_email))
    error_message = "Notification email must be a valid email address or empty string."
  }
}

# ============================================================================
# FEATURE FLAGS
# ============================================================================

variable "enable_deletion_protection" {
  description = "Enable deletion protection for RDS instance"
  type        = bool
  default     = true
}

variable "enable_multi_az" {
  description = "Enable Multi-AZ deployment for RDS"
  type        = bool
  default     = false
}

variable "enable_performance_insights" {
  description = "Enable Performance Insights for RDS"
  type        = bool
  default     = true
} 