# Outputs for Stock Analyzer Infrastructure
# Organized by layer following clean architecture principles

# ============================================================================
# CORE APPLICATION OUTPUTS
# ============================================================================

output "application_url" {
  description = "Public URL of the application load balancer"
  value       = module.application.application_url
}

output "health_check_url" {
  description = "Health check endpoint URL"
  value       = "${module.application.application_url}/health"
}

output "api_base_url" {
  description = "Base URL for API endpoints"
  value       = "${module.application.application_url}/api/v1"
}

# ============================================================================
# NETWORKING OUTPUTS
# ============================================================================

output "vpc_id" {
  description = "ID of the VPC"
  value       = module.networking.vpc_id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC"
  value       = module.networking.vpc_cidr_block
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = module.networking.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = module.networking.private_subnet_ids
}

output "database_subnet_ids" {
  description = "IDs of the database subnets"
  value       = module.networking.database_subnet_ids
}

# ============================================================================
# DATABASE OUTPUTS
# ============================================================================

output "database_endpoint" {
  description = "RDS instance endpoint"
  value       = module.database.database_host
  sensitive   = false
}

output "database_port" {
  description = "RDS instance port"
  value       = module.database.database_port
}

output "database_name" {
  description = "Database name"
  value       = module.database.database_name
}

output "database_username" {
  description = "Database master username"
  value       = module.database.database_username
  sensitive   = false
}

output "database_password_secret_arn" {
  description = "ARN of the secret containing the database password"
  value       = module.database.database_password_secret_arn
  sensitive   = true
}

# ============================================================================
# APPLICATION LAYER OUTPUTS
# ============================================================================

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = module.application.ecs_cluster_name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = module.application.ecs_service_name
}

output "task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = module.application.task_definition_arn
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.application.alb_dns_name
}

output "alb_zone_id" {
  description = "Hosted zone ID of the Application Load Balancer"
  value       = module.application.alb_zone_id
}

# ============================================================================
# SECURITY OUTPUTS
# ============================================================================

output "waf_web_acl_arn" {
  description = "ARN of the WAF Web ACL (if enabled)"
  value       = var.enable_waf ? module.security.waf_web_acl_arn : null
}

output "alb_security_group_id" {
  description = "ID of the ALB security group"
  value       = module.application.alb_security_group_id
}

output "app_security_group_id" {
  description = "ID of the application security group"
  value       = module.application.app_security_group_id
}

# ============================================================================
# MONITORING OUTPUTS
# ============================================================================

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = module.application.cloudwatch_log_group_name
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts (if email provided)"
  value       = var.notification_email != "" ? module.monitoring.sns_topic_arn : null
}

# ============================================================================
# UTILITY OUTPUTS
# ============================================================================

output "migration_task_definition_arn" {
  description = "ARN of the migration task definition"
  value       = module.migration.task_definition_arn
}

# ============================================================================
# RESOURCE IDENTIFIERS
# ============================================================================

output "resource_tags" {
  description = "Common tags applied to all resources"
  value = {
    Application = var.app_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "stock-analyzer"
  }
}

output "app_name_full" {
  description = "Full application name including environment"
  value       = "${var.app_name}-${var.environment}"
}

# ============================================================================
# DEPLOYMENT INFORMATION
# ============================================================================

output "deployment_info" {
  description = "Key deployment information"
  value = {
    application_url    = module.application.application_url
    health_check_url   = "${module.application.application_url}/health"
    api_base_url       = "${module.application.application_url}/api/v1"
    environment        = var.environment
    aws_region         = var.aws_region
    vpc_id            = module.networking.vpc_id
    ecs_cluster       = module.application.ecs_cluster_name
    ecs_service       = module.application.ecs_service_name
    database_endpoint = module.database.database_host
  }
} 