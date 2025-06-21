output "lambda_functions" {
  description = "Map of Lambda functions"
  value       = aws_lambda_function.functions
}

output "lambda_function_arns" {
  description = "ARNs of the Lambda functions"
  value = {
    for name, func in aws_lambda_function.functions : name => func.arn
  }
}

output "lambda_function_names" {
  description = "Names of the Lambda functions"
  value = {
    for name, func in aws_lambda_function.functions : name => func.function_name
  }
}

output "lambda_function_invoke_arns" {
  description = "Invoke ARNs of the Lambda functions"
  value = {
    for name, func in aws_lambda_function.functions : name => func.invoke_arn
  }
}

output "lambda_role_arn" {
  description = "ARN of the Lambda execution role"
  value       = aws_iam_role.lambda.arn
}

output "lambda_role_name" {
  description = "Name of the Lambda execution role"
  value       = aws_iam_role.lambda.name
}

output "api_function_url" {
  description = "Function URL for the API Lambda"
  value       = aws_lambda_function_url.api.function_url
}

output "lambda_layer_arn" {
  description = "ARN of the common Lambda layer"
  value       = aws_lambda_layer_version.common.arn
}

output "cloudwatch_log_groups" {
  description = "CloudWatch log groups for Lambda functions"
  value = {
    for name, log_group in aws_cloudwatch_log_group.lambda_logs : name => log_group.name
  }
} 