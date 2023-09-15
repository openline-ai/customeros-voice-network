module "homer" {
  source = "../../../../../modules/ec2/openline-services/homer"
  environment = var.environment
  ec2_ssh_key_name = var.ec2_ssh_key_name
  aws_region = var.aws_region
  openline_hosted_zone_id = var.openline_hosted_network_zone_id
  vpc_id = var.vpc_id
  homer_instance_type = var.homer_instance_type
  public_subnets_ids = var.public_subnets_id
  openline_certificate = var.openline_network_certificate
}