module "asterisk" {
  source = "../../../../../modules/ec2/openline-services/asterisk"
  environment = var.environment
  ec2_ssh_key_name = var.ec2_ssh_key_name
  aws_region = var.aws_region
  openline_hosted_zone_id = var.openline_hosted_network_zone_id
  vpc_id = var.vpc_id
  asterisk_instance_type = var.asterisk_instance_type
  public_subnets_ids = var.public_subnets_id
  cloudwatch_push_iam = var.cloudwatch_push_iam

}