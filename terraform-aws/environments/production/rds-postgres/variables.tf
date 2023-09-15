variable "aws_region" {}

variable "environment" {}

variable "vpc_id" {}
variable "cidr" {}

variable "private_subnets_id" {}

variable "rds_postgres_type" {}

variable "rds_postgres_major_version" {}
variable "rds_postgres_version" {}
variable "rds_postgres_autoscale" {}
variable "rds_postgres_autoscale_min_capacity" {}
variable "rds_postgres_autoscale_max_capacity" {}

variable "sns_notification_arn" {}
variable "freeable_memory" {}
variable "freeable_storage_space" {}
variable "read_iops" {}
variable "write_iops" {}