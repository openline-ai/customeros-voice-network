provider "aws" {
  region = var.aws_region
}

module "vpc" {
  source = "../../../modules/vpc"

  environment = var.environment

  ssh_key_name = "${var.ec2_ssh_key_name}"

  aws_region      = var.aws_region
  cidr            = var.cidr
  azs             = var.azs
  private_subnets = var.private_subnets
  public_subnets  = var.public_subnets
  openline_hosted_zone_id = var.openline_hosted_zone_id
  ssh_jump_allow_lists = var.ssh_jump_allow_lists
}