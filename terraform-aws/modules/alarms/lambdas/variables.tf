variable "environment" {}
variable "aws_region" {}

variable "errors-openline-ai-sender" {
  description = "Errors in openline-ai-sender lambda"
  type        = number
  default     = 1
}

variable "sns_notification_arn" {}