provider "aws" {
  region = var.aws_region
}

module "rds_postgres" {
  source = "../../../modules/rds-postgres"

  environment = var.environment

  vpc_id             = var.vpc_id
  cidr               = var.cidr
  private_subnets_id = var.private_subnets_id
  rds_postgres_type  = var.rds_postgres_type

  rds_postgres_major_version          = var.rds_postgres_major_version
  rds_postgres_version                = var.rds_postgres_version
  rds_postgres_autoscale              = var.rds_postgres_autoscale
  rds_postgres_autoscale_min_capacity = var.rds_postgres_autoscale_min_capacity
  rds_postgres_autoscale_max_capacity = var.rds_postgres_autoscale_max_capacity
  aws_region                          = var.aws_region
  sns_notification_arn                = var.sns_notification_arn
  create_alarm_freeable_memory        = var.freeable_memory
  create_alarm_free_local_storage     = var.freeable_storage_space
  create_alarm_read_iops              = var.read_iops
  create_alarm_write_iops             = var.write_iops
}