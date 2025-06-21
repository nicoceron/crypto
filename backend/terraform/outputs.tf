output "api_gateway_url" {
  description = "URL of the API Gateway"
  value       = module.api_gateway.api_gateway_url
}

output "api_gateway_id" {
  description = "ID of the API Gateway"
  value       = module.api_gateway.api_gateway_id
}

output "lambda_functions" {
  description = "Information about Lambda functions"
  value = {
    for name, func in module.lambda.lambda_functions : name => {
      function_name = func.function_name
      function_arn  = func.arn
      invoke_arn    = func.invoke_arn
    }
  }
}

# Database outputs removed - using external CockroachDB

output "vpc_id" {
  description = "ID of the VPC"
  value       = module.networking.vpc_id
}

output "app_subnet_ids" {
  description = "IDs of the application subnets"
  value       = module.networking.app_subnet_ids
}

# Database subnet outputs removed - using external CockroachDB

output "s3_bucket_name" {
  description = "Name of the S3 bucket for Lambda deployments"
  value       = aws_s3_bucket.lambda_deployments.bucket
}

output "region" {
  description = "AWS region"
  value       = data.aws_region.current.name
}

output "account_id" {
  description = "AWS account ID"
  value       = data.aws_caller_identity.current.account_id
}

# Frontend outputs
output "frontend_url" {
  description = "URL of the deployed frontend application"
  value       = module.frontend.frontend_url
}

output "frontend_s3_bucket" {
  description = "Name of the S3 bucket hosting the frontend"
  value       = module.frontend.s3_bucket_name
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = module.frontend.cloudfront_distribution_id
}

output "cloudfront_domain_name" {
  description = "Domain name of the CloudFront distribution"
  value       = module.frontend.cloudfront_domain_name
} 