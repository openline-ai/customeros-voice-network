provider "aws" {
  region = var.aws_region
}

locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
    Environment = var.environment
    Service    = "Kamailio"
    CostIdentifier = "Voice Network"
  }
}

data "aws_vpc" "selected" {
  id = var.vpc_id
}

module "alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 6.0"

  name = "kamailio-${var.environment}-ws"

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
        port                = "8080"
        unhealthy_threshold = 2
        timeout = 2
        interval = 5
      }

      targets = {
        my_kamailio_ws1 = {
          target_id = aws_instance.instance1.id
          port      = 8080
        }
        my_kamailio_ws2 = {
          target_id = aws_instance.instance2.id
          port      = 8080
        }
      }
    }
  ]

  tags = local.common_tags
}

module "nlb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 6.0"

  name = "kamailio-${var.environment}-sip"

  load_balancer_type = "network"

  vpc_id  = var.vpc_id
  subnets = var.public_subnets_ids
  #security_groups = [aws_security_group.lb.id]

  http_tcp_listeners = [
    {
      port               = 5060
      protocol           = "TCP_UDP"
      target_group_index = 0
    }
  ]

  target_groups = [
    {
      name_prefix      = "kam"
      backend_protocol = "TCP_UDP"
      backend_port     = 5060
      target_type      = "instance"

      health_check = {
        protocol            = "TCP"
        port                = "8080"
        unhealthy_threshold = 2
        timeout = 2
        interval = 5
      }
      
      stickiness = {
        enabled = true
        type = "source_ip"
      }

      targets = {
        my_kamailio_sip1 = {
          target_id = aws_instance.instance1.id
          port      = 5060
        }
        my_kamailio_sip2 = {
          target_id = aws_instance.instance2.id
          port      = 5060
        }
      }
    }
  ]

  tags = local.common_tags
}


data "aws_ami" "ami" {
  most_recent      = true
  owners           = ["self"]

  filter {
    name   = "name"
    values = ["openline-voice-kamailio_${var.environment}"]
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

    from_port = 8080
    to_port   = 8080
    protocol  = "tcp"
  }
  ingress {
    cidr_blocks = [data.aws_vpc.selected.cidr_block]

    from_port = 5090
    to_port   = 5090
    protocol  = "udp"
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]

    from_port = 5060
    to_port   = 5060
    protocol  = "udp"
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]

    from_port = 5060
    to_port   = 5060
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

resource "aws_security_group" "lb" {
  vpc_id  = var.vpc_id

  ingress {
    cidr_blocks = ["0.0.0.0/0"]

    from_port = 443
    to_port   = 443
    protocol  = "tcp"
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]

    from_port = 5060
    to_port   = 5060
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

resource "aws_instance" "instance1" {
  ami             = data.aws_ami.ami.id
  instance_type   = var.kamailio_instance_type
  vpc_security_group_ids = [aws_security_group.sg.id]
  key_name = var.ec2_ssh_key_name
  iam_instance_profile =  var.cloudwatch_push_iam
  associate_public_ip_address = true
  tags = merge(
    tomap({Name = "Kamailio Server ${var.environment} 1"}),
    local.common_tags)
  subnet_id = var.public_subnets_ids[0]
    metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }
}

resource "aws_instance" "instance2" {
  ami             = data.aws_ami.ami.id
  instance_type   = var.kamailio_instance_type
  vpc_security_group_ids = [aws_security_group.sg.id]
  key_name = var.ec2_ssh_key_name
  iam_instance_profile =  var.cloudwatch_push_iam
  associate_public_ip_address = true

  tags = merge(
    tomap({Name = "Kamailio Server ${var.environment} 2"}),
    local.common_tags)
  subnet_id = var.public_subnets_ids[0]
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }
}


resource "aws_route53_record" "ec2_kamailio_wss_cname_record" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "kamailio-${var.environment}"
  records = [module.alb.lb_dns_name]
  ttl     = 300
}

resource "aws_route53_record" "ec2_kamailio_sip_udp_srv_record" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "SRV"
  name    = "_sip._udp.kamailio-${var.environment}"
  records = ["0 10 5060 ${module.nlb.lb_dns_name}"]
  ttl     = 300
}

resource "aws_route53_record" "ec2_kamailio_sip_tcp_srv_record" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "SRV"
  name    = "_sip._tcp.kamailio-${var.environment}"
  records = ["0 10 5060 ${module.nlb.lb_dns_name}"]
  ttl     = 300
}


resource "aws_route53_record" "ec2_kamailio_sip_udp_cname_record" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "sip.kamailio-${var.environment}"
  records = [module.nlb.lb_dns_name]
  ttl     = 300
}

resource "aws_route53_record" "ec2_kamailio_dmq_a_record" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "A"
  name    = "kamailio-dmq-${var.environment}"
  records = [aws_instance.instance1.private_ip, aws_instance.instance2.private_ip]
  ttl     = 300
}

resource "aws_route53_record" "instance1" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "A"
  name    = "kamailio1.kamailio-${var.environment}"
  records = [aws_instance.instance1.public_ip]
  ttl     = 300
}

resource "aws_route53_record" "instance2" {
  depends_on = [aws_instance.instance1, aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "A"
  name    = "kamailio2.kamailio-${var.environment}"
  records = [aws_instance.instance2.public_ip]
  ttl     = 300
}
