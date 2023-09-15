provider "aws" {
  region = var.aws_region
}

locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
    Service = "Homer ${var.environment}"
    CostIdentifier = "Voice Network"
  }
}

data "aws_vpc" "selected" {
  id = var.vpc_id
}

data "aws_ami" "ami" {
  most_recent      = true
  owners           = ["self"]

  filter {
    name   = "name"
    values = ["openline-voice-homer"]
  }
  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
module "alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 6.0"

  name = "homer-${var.environment}-alb"

  load_balancer_type = "application"

  vpc_id  = var.vpc_id
  subnets = var.public_subnets_ids
  security_groups = [aws_security_group.lb.id]

  https_listeners = [
    {
      port               = 443
      protocol           = "HTTPS"
      certificate_arn    = var.openline_certificate
      target_group_index = 0
    }
  ]

  target_groups = [
    {
      name_prefix      = "kam"
      backend_protocol = "HTTP"
      backend_port     = 8080
      target_type      = "instance"

      health_check = {
        protocol            = "HTTP"
        port                = "9080"
        unhealthy_threshold = 2
        timeout = 2
        interval = 5
      }

      targets = {
        my_homer = {
          target_id = aws_instance.instance.id
          port      = 9080
        }
      }
    }
  ]

  tags = local.common_tags
  
}

resource "aws_security_group" "lb" {
  vpc_id  = var.vpc_id

  ingress {
    cidr_blocks = ["0.0.0.0/0"]

    from_port = 443
    to_port   = 443
    protocol  = "tcp"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags

}

data "aws_instance" "jump" {
  filter {
    name   = "tag:Service"
    values = ["Jump"]
  }
  filter {
    name   = "tag:Environment"
    values = [var.environment]
  }
}

resource "aws_security_group" "sg" {
  vpc_id  = var.vpc_id

  ingress {
    cidr_blocks = ["${data.aws_instance.jump.private_ip}/32", "${data.aws_instance.jump.public_ip}/32"]

    from_port = 22
    to_port   = 22
    protocol  = "tcp"
  }

  ingress {
    cidr_blocks = [data.aws_vpc.selected.cidr_block]

    from_port = 9060
    to_port   = 9060
    protocol  = "udp"
  }
  ingress {
    cidr_blocks = [data.aws_vpc.selected.cidr_block]

    from_port = 9060
    to_port   = 9060
    protocol  = "tcp"
  }
  ingress {
    cidr_blocks = [data.aws_vpc.selected.cidr_block]

    from_port = 9080
    to_port   = 9080
    protocol  = "tcp"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "instance" {
  ami             = data.aws_ami.ami.id
  instance_type   = var.homer_instance_type
  vpc_security_group_ids = [aws_security_group.sg.id]
  key_name = var.ec2_ssh_key_name

  tags = merge(
    tomap({Name = "Homer Server ${var.environment}"}),
    local.common_tags)
  subnet_id = var.public_subnets_ids[0]

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }
}

resource "aws_route53_record" "homer_lb_cname" {
  depends_on = [aws_instance.instance]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "homer-${var.environment}"
  records = [module.alb.lb_dns_name]
  ttl     = 300
}

resource "aws_route53_record" "homer_instance_cname" {
  depends_on = [aws_instance.instance]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "homer-internal-${var.environment}"
  records = [aws_instance.instance.private_dns]
  ttl     = 300
}
