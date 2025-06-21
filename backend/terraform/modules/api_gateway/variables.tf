variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
}

variable "lambda_functions" {
  description = "Map of Lambda functions"
  type = map(object({
    function_name = string
    arn          = string
    invoke_arn   = string
  }))
}

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
} 