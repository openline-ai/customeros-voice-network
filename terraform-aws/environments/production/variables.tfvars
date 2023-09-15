aws_account_id = ""
environment    = "production"
aws_region     = "eu-west-1"

#VPC
azs              = [""]
cidr             = ""
private_subnets  = [""]
public_subnets   = [""]
ec2_ssh_key_name = ""

#VPC after creation
vpc_id               = ""
vpc_arn              = ""
cloudwatch_push_iam  = ""
private_subnets_id   = ["", "", "",]
public_subnets_id    = ["", "", "",]
ssh_jump_allow_lists = [
  {
    subnets     = ["your.public.ip.address/32"]
    description = ""
  },
]

//rds postgres
rds_postgres_major_version          = "14"
rds_postgres_version                = "14.5"
rds_postgres_type                   = "db.r5.large" //2 CPU, 16GB RAM
rds_postgres_autoscale              = true
rds_postgres_autoscale_min_capacity = 1
rds_postgres_autoscale_max_capacity = 3

rds_postgres_cluster_password_arn = ""

#Route53
openline_hosted_network_zone_id = ""
openline_network_certificate    = ""

# notifications
sns_notification_arn = ""

#EC2 - Kamailio
kamailio_instance_type = "c6a.large"
kamailio_dmq_domain    = ""
homer_internal_ip      = ""

#EC2 - Asterisk
asterisk_instance_type = "c6a.large"

#EC2 - Homer
homer_instance_type = "c6a.large"

#Alarms
freeable_memory        = true
freeable_storage_space = true
read_iops              = true
write_iops             = true

#redis
node_type     = "db.t4g.small"
custom_domain = ""