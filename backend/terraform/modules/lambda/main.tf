# Data source for current region and account
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# IAM role for Lambda functions
resource "aws_iam_role" "lambda" {
  name = "${var.project_name}-${var.environment}-lambda-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
  
  tags = var.common_tags
}

# IAM policy for Lambda functions
resource "aws_iam_role_policy" "lambda" {
  name = "${var.project_name}-${var.environment}-lambda-policy"
  role = aws_iam_role.lambda.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Effect = "Allow"
        Action = [
          "ec2:CreateNetworkInterface",
          "ec2:DescribeNetworkInterfaces",
          "ec2:DeleteNetworkInterface"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = "*"
      }
    ]
  })
}

# Attach basic execution role policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Attach VPC access execution role policy
resource "aws_iam_role_policy_attachment" "lambda_vpc" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# CloudWatch log groups for Lambda functions
resource "aws_cloudwatch_log_group" "lambda_logs" {
  for_each = local.lambda_functions
  
  name              = "/aws/lambda/${var.project_name}-${var.environment}-${each.key}"
  retention_in_days = var.environment == "prod" ? 14 : 7
  
  tags = var.common_tags
}

# Local values for Lambda functions configuration
locals {
  lambda_functions = {
    # API functions - one per major route group
    api = {
      handler     = "bootstrap"
      runtime     = "provided.al2"
      timeout     = 120  # Increased from 30 to 120 seconds
      memory_size = 1024  # Increased from 512 to 1024 MB
      description = "Main API handler for stock analyzer"
    }
    
    # Background job functions
    ingestion = {
      handler     = "bootstrap"
      runtime     = "provided.al2"
      timeout     = 900  # 15 minutes for data ingestion
      memory_size = 1024
      description = "Data ingestion service"
    }
    
    # Scheduled functions
    scheduler = {
      handler     = "bootstrap"
      runtime     = "provided.al2"
      timeout     = 60
      memory_size = 256
      description = "Scheduled tasks handler"
    }
  }
  
  # Common environment variables for all Lambda functions
  common_environment_variables = {
    ENVIRONMENT       = var.environment
    DATABASE_URL      = var.database_url
    ALPACA_API_KEY    = var.alpaca_api_key
    ALPACA_API_SECRET = var.alpaca_api_secret
    STOCK_API_URL     = var.stock_api_url
    STOCK_API_TOKEN   = var.stock_api_token
    LOG_LEVEL         = var.environment == "prod" ? "info" : "debug"
  }
}

# Lambda functions
resource "aws_lambda_function" "functions" {
  for_each = local.lambda_functions
  
  function_name = "${var.project_name}-${var.environment}-${each.key}"
  role         = aws_iam_role.lambda.arn
  
  # Deployment package configuration
  filename         = "${path.module}/lambda-placeholder.zip"
  source_code_hash = data.archive_file.lambda_placeholder.output_base64sha256
  
  # Runtime configuration
  runtime = each.value.runtime
  handler = each.value.handler
  timeout = each.value.timeout
  
  # Memory and compute configuration
  memory_size = each.value.memory_size
  
  # Publish versions for API function to enable provisioned concurrency
  publish = each.key == "api" ? true : false
  
  # Network configuration
  vpc_config {
    subnet_ids         = var.app_subnet_ids
    security_group_ids = [var.app_security_group_id]
  }
  
  # Environment variables
  environment {
    variables = merge(local.common_environment_variables, {
      FUNCTION_TYPE = each.key
    })
  }
  
  # Logging configuration
  depends_on = [aws_cloudwatch_log_group.lambda_logs]
  
  tags = merge(var.common_tags, {
    Name        = "${var.project_name}-${var.environment}-${each.key}"
    Description = each.value.description
  })
  
  lifecycle {
    ignore_changes = [
      filename,
      source_code_hash
    ]
  }
}

# Lambda alias for API function to enable stable versioning (keeping for future use)
resource "aws_lambda_alias" "api_alias" {
  name             = "live"
  function_name    = aws_lambda_function.functions["api"].function_name
  function_version = aws_lambda_function.functions["api"].version
}

# Note: Provisioned concurrency removed due to AWS account concurrency limits
# Account doesn't have enough unreserved concurrency (needs min 10, even 2 provisioned hits limit)

# Placeholder zip file for initial deployment
data "archive_file" "lambda_placeholder" {
  type        = "zip"
  output_path = "${path.module}/lambda-placeholder.zip"
  
  source {
    content  = "placeholder"
    filename = "bootstrap"
  }
}

# Lambda function URLs for direct invocation (optional)
resource "aws_lambda_function_url" "api" {
  function_name      = aws_lambda_function.functions["api"].function_name
  authorization_type = "NONE"
  
  cors {
    allow_credentials = false
    allow_origins     = ["*"]
    allow_methods     = ["*"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["date", "keep-alive"]
    max_age          = 86400
  }
}

# EventBridge rule for scheduled ingestion
resource "aws_cloudwatch_event_rule" "ingestion_schedule" {
  name                = "${var.project_name}-${var.environment}-ingestion-schedule"
  description         = "Trigger ingestion Lambda function"
  schedule_expression = "rate(4 hours)"  # Run every 4 hours
  
  tags = var.common_tags
}

# EventBridge target for ingestion
resource "aws_cloudwatch_event_target" "ingestion_target" {
  rule      = aws_cloudwatch_event_rule.ingestion_schedule.name
  target_id = "IngestionLambdaTarget"
  arn       = aws_lambda_function.functions["ingestion"].arn
}

# Lambda permission for EventBridge
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.functions["ingestion"].function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.ingestion_schedule.arn
}

# Lambda layers for common dependencies
resource "aws_lambda_layer_version" "common" {
  filename   = "${path.module}/layer-placeholder.zip"
  layer_name = "${var.project_name}-${var.environment}-common-layer"
  
  compatible_runtimes = ["provided.al2"]
  description         = "Common dependencies layer"
  
  lifecycle {
    ignore_changes = [
      filename,
      source_code_hash
    ]
  }
}

# Placeholder layer zip file
data "archive_file" "layer_placeholder" {
  type        = "zip"
  output_path = "${path.module}/layer-placeholder.zip"
  
  source {
    content  = "placeholder"
    filename = "lib/placeholder"
  }
} 