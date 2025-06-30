# API Gateway REST API
resource "aws_api_gateway_rest_api" "main" {
  name        = "${var.project_name}-${var.environment}-api"
  description = "API Gateway for ${var.project_name} ${var.environment} environment"
  
  endpoint_configuration {
    types = ["REGIONAL"]
  }
  
  tags = var.common_tags
}

# API Gateway resources and methods

# Health check resource
resource "aws_api_gateway_resource" "health" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_rest_api.main.root_resource_id
  path_part   = "health"
}

resource "aws_api_gateway_method" "health" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.health.id
  http_method   = "GET"
  authorization = "NONE"
}

# API v1 resource
resource "aws_api_gateway_resource" "api" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_rest_api.main.root_resource_id
  path_part   = "api"
}

resource "aws_api_gateway_resource" "v1" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.api.id
  path_part   = "v1"
}

# Ratings resources
resource "aws_api_gateway_resource" "ratings" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "ratings"
}

resource "aws_api_gateway_method" "ratings_get" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.ratings.id
  http_method   = "GET"
  authorization = "NONE"
}

# Ratings by ticker resource
resource "aws_api_gateway_resource" "ratings_ticker" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.ratings.id
  path_part   = "{ticker}"
}

resource "aws_api_gateway_method" "ratings_ticker_get" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.ratings_ticker.id
  http_method   = "GET"
  authorization = "NONE"
}

# Recommendations resource
resource "aws_api_gateway_resource" "recommendations" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "recommendations"
}

resource "aws_api_gateway_method" "recommendations_get" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.recommendations.id
  http_method   = "GET"
  authorization = "NONE"
}

# Stocks resource
resource "aws_api_gateway_resource" "stocks" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "stocks"
}

resource "aws_api_gateway_resource" "stocks_symbol" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.stocks.id
  path_part   = "{symbol}"
}

# Stock price resource
resource "aws_api_gateway_resource" "stocks_price" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.stocks_symbol.id
  path_part   = "price"
}

resource "aws_api_gateway_method" "stocks_price_get" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.stocks_price.id
  http_method   = "GET"
  authorization = "NONE"
}

# Stock logo resource
resource "aws_api_gateway_resource" "stocks_logo" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.stocks_symbol.id
  path_part   = "logo"
}

resource "aws_api_gateway_method" "stocks_logo_get" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.stocks_logo.id
  http_method   = "GET"
  authorization = "NONE"
}

# Ingest resource
resource "aws_api_gateway_resource" "ingest" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "ingest"
}

resource "aws_api_gateway_method" "ingest_post" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.ingest.id
  http_method   = "POST"
  authorization = "NONE"
}

# Lambda integrations for all methods
locals {
  lambda_integrations = {
    health                = aws_api_gateway_method.health
    ratings_get          = aws_api_gateway_method.ratings_get
    ratings_ticker_get   = aws_api_gateway_method.ratings_ticker_get
    recommendations_get  = aws_api_gateway_method.recommendations_get
    stocks_price_get     = aws_api_gateway_method.stocks_price_get
    stocks_logo_get      = aws_api_gateway_method.stocks_logo_get
    ingest_post          = aws_api_gateway_method.ingest_post
  }
}

# Lambda integrations
resource "aws_api_gateway_integration" "lambda" {
  for_each = local.lambda_integrations
  
  rest_api_id             = aws_api_gateway_rest_api.main.id
  resource_id             = each.value.resource_id
  http_method             = each.value.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.lambda_functions["api"].invoke_arn
}

# Lambda permissions for API Gateway
resource "aws_lambda_permission" "apigw" {
  for_each = local.lambda_integrations
  
  statement_id  = "AllowExecutionFromAPIGateway-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_functions["api"].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.main.execution_arn}/*/*"
}

# CORS configuration for all resources
resource "aws_api_gateway_method" "cors" {
  for_each = {
    health                = aws_api_gateway_resource.health
    ratings              = aws_api_gateway_resource.ratings
    ratings_ticker       = aws_api_gateway_resource.ratings_ticker
    recommendations      = aws_api_gateway_resource.recommendations
    stocks_price         = aws_api_gateway_resource.stocks_price
    stocks_logo          = aws_api_gateway_resource.stocks_logo
    ingest               = aws_api_gateway_resource.ingest
  }
  
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = each.value.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "cors" {
  for_each = aws_api_gateway_method.cors
  
  rest_api_id = aws_api_gateway_rest_api.main.id
  resource_id = each.value.resource_id
  http_method = each.value.http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "cors" {
  for_each = aws_api_gateway_method.cors
  
  rest_api_id = aws_api_gateway_rest_api.main.id
  resource_id = each.value.resource_id
  http_method = each.value.http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Expose-Headers" = true
    "method.response.header.Access-Control-Allow-Credentials" = true
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "cors" {
  for_each = aws_api_gateway_integration.cors
  
  rest_api_id = aws_api_gateway_rest_api.main.id
  resource_id = each.value.resource_id
  http_method = each.value.http_method
  status_code = aws_api_gateway_method_response.cors[each.key].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type, Authorization, X-Requested-With, Accept, Origin, X-Api-Key, X-Amz-Date, X-Amz-Security-Token'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET, POST, PUT, DELETE, OPTIONS, HEAD'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${var.frontend_url}'"
    "method.response.header.Access-Control-Expose-Headers" = "'Content-Length, Content-Type'"
    "method.response.header.Access-Control-Allow-Credentials" = "'false'"
    "method.response.header.Access-Control-Max-Age" = "'86400'"
  }
}

# API Gateway deployment
resource "aws_api_gateway_deployment" "main" {
  depends_on = [
    aws_api_gateway_integration.lambda,
    aws_api_gateway_integration.cors
  ]
  
  rest_api_id = aws_api_gateway_rest_api.main.id
  stage_name  = var.environment
  
  lifecycle {
    create_before_destroy = true
  }
  
  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_rest_api.main.body,
      aws_api_gateway_integration.lambda,
      aws_api_gateway_integration.cors
    ]))
  }
}

# API Gateway stage configuration
resource "aws_api_gateway_stage" "main" {
  deployment_id = aws_api_gateway_deployment.main.id
  rest_api_id   = aws_api_gateway_rest_api.main.id
  stage_name    = var.environment
  
  # Enable caching for better performance
  cache_cluster_enabled = var.environment == "prod" ? true : false
  cache_cluster_size    = var.environment == "prod" ? "0.5" : null
  
  # Throttling is configured via method settings below
  
  # Enable access logging with proper CloudWatch setup
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      extendedRequestId = "$context.extendedRequestId"
      ip             = "$context.sourceIp"
      caller         = "$context.caller"
      user           = "$context.user"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      resourcePath   = "$context.resourcePath"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }
  
  tags = var.common_tags
}

# CloudWatch log group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.project_name}-${var.environment}"
  retention_in_days = var.environment == "prod" ? 14 : 7
  
  tags = var.common_tags
}

# IAM role for API Gateway CloudWatch logging
resource "aws_iam_role" "api_gateway_cloudwatch" {
  name = "${var.project_name}-${var.environment}-api-gateway-cloudwatch-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })
  
  tags = var.common_tags
}

resource "aws_iam_role_policy_attachment" "api_gateway_cloudwatch" {
  role       = aws_iam_role.api_gateway_cloudwatch.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
}

# API Gateway account settings for CloudWatch logging
resource "aws_api_gateway_account" "main" {
  cloudwatch_role_arn = aws_iam_role.api_gateway_cloudwatch.arn
}

# API Gateway method settings for logging
resource "aws_api_gateway_method_settings" "all" {
  depends_on = [aws_api_gateway_account.main]
  
  rest_api_id = aws_api_gateway_rest_api.main.id
  stage_name  = aws_api_gateway_stage.main.stage_name
  method_path = "*/*"
  
  settings {
    logging_level      = var.environment == "prod" ? "INFO" : "ERROR"
    data_trace_enabled = var.environment != "prod"
    metrics_enabled    = true
    throttling_rate_limit  = 5000  # Increased from 1000 to 5000 requests/second
    throttling_burst_limit = 10000  # Increased from 2000 to 10000 burst
  }
} 