provider "aws" {
  region = var.aws_region
}

locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
  }
}

data "aws_secretsmanager_secret" "aws_database_postgres_password" {
  arn = var.rds_postgres_cluster_password_arn
}

data "aws_secretsmanager_secret_version" "current" {
  secret_id = data.aws_secretsmanager_secret.aws_database_postgres_password.id
}
