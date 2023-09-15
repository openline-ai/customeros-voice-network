locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
  }
}

resource "aws_s3_bucket" "terraform_state_bucket" {
  bucket = "${var.environment}-terraform-state"
  force_destroy = true
  tags = merge(
    local.common_tags,
    {
      CostIdentifier = "${var.environment} terraform s3 state bucket"
    }
  )
}