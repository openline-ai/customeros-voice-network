terraform {
  required_version = ">= 0.13.1"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 2.49"
    }
  }

  backend "s3" {
    bucket         = "voice-network-terraform-state"
    region         = "eu-west-1"
    dynamodb_table = "terraform-state-lock-voice-network-ec2-homer"
  }
}
