locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
  }
  aws_sns_topic_arn = var.sns_notification_arn
}

data "aws_lambda_functions" "all_lambdas" {
}

resource "aws_cloudwatch_metric_alarm" "lambdas_error_alarm" {
  count               = length(data.aws_lambda_functions.all_lambdas.function_names)
  alarm_name          = join("-", [data.aws_lambda_functions.all_lambdas.function_names[count.index], "alarm"])
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "1"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "60"
  statistic           = "SampleCount"
  threshold           = 1
  alarm_description   = "Triggered when Lambda function errors occur"
  alarm_actions       = [local.aws_sns_topic_arn]
  ok_actions          = [local.aws_sns_topic_arn]

  dimensions = {
    LambdaFunctionArn = data.aws_lambda_functions.all_lambdas.function_arns[count.index]
  }
}