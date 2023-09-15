variable "aws_region" {}

variable "environment" {}
variable "ec2_ssh_key_name" {}

variable "azs" { }

variable "cidr" {}

variable "private_subnets" {}
variable "public_subnets" {}

variable "openline_hosted_zone_id" {}

variable "ssh_jump_allow_lists" {
  type = list(object({
    subnets = list(string)
    description = string
  }))
}