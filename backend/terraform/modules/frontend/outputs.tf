output "s3_bucket_name" {
  description = "Name of the S3 bucket hosting the frontend"
  value       = aws_s3_bucket.frontend.bucket
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket hosting the frontend"
  value       = aws_s3_bucket.frontend.arn
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.id
}

output "cloudfront_domain_name" {
  description = "Domain name of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "frontend_url" {
  description = "URL of the deployed frontend application"
  value       = "https://${aws_cloudfront_distribution.frontend.domain_name}"
}

output "s3_bucket_website_endpoint" {
  description = "S3 bucket website endpoint"
  value       = aws_s3_bucket.frontend.bucket_regional_domain_name
} 