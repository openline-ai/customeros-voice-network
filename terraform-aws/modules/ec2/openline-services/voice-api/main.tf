provider "aws" {
  region = var.aws_region
}

locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
    Environment = var.environment
    Service    = "Voice API"
    CostIdentifier = "Voice Network"
  }
}

data "aws_vpc" "selected" {
  id = var.vpc_id
}

data "aws_vpc" "guest" {
  id = var.vpc_id
}

module "alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 6.0"

  name = "voice-api-${var.environment}-alb"

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
      name_prefix      = "vapi"
      backend_protocol = "HTTP"
      backend_port     = 8080
      target_type      = "instance"

      health_check = {
        protocol            = "HTTP"
        port                = "8080"
        unhealthy_threshold = 2
        timeout = 2
        interval = 30
      }

      targets = {
        my_vapi_1 = {
          target_id = aws_instance.instance1.id
          port      = 8080
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
    values = ["voice-api-server-ami_${var.environment}"]
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
    cidr_blocks = [data.aws_vpc.guest.cidr_block]


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

module "cloudwatch_iam_profile" {
  source = "../../../iam_role/push_logs_to_cloudwatch"
  environment = var.environment
  service = "voice-api"

}

module "s3_voicemail_bucket" {
  source = "../../../iam_role/s3_bucket"
  environment = var.environment
  service = "voice-api"
  s3_bucket_name = var.aws_s3_bucket
}

data "aws_iam_policy_document" "role_policy" {
  statement {
    actions = [
      "sts:AssumeRole",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "instance_role" {
  name = "voice-api-access-role"
  assume_role_policy = data.aws_iam_policy_document.role_policy.json
}


resource "aws_iam_instance_profile" "instance_profile" {
  name = "${var.environment}-voice-api-access-profile"
  role = aws_iam_role.instance_role.name
}

resource "aws_iam_role_policy_attachment" "policy_attachment_cloudwatch" {
  policy_arn = module.cloudwatch_iam_profile.policy_arn
  role       = aws_iam_role.instance_role.name
}

resource "aws_iam_role_policy_attachment" "policy_attachment_s3" {
  policy_arn = module.s3_voicemail_bucket.policy_arn
  role       = aws_iam_role.instance_role.name
}

resource "aws_instance" "instance1" {
  ami             = data.aws_ami.ami.id
  instance_type   = var.voice_api_instance_type
  vpc_security_group_ids = [aws_security_group.sg.id]
  key_name = var.ec2_ssh_key_name
  iam_instance_profile =  aws_iam_instance_profile.instance_profile.name
  associate_public_ip_address = true
  tags = merge(
    tomap({Name = "Voice API ${var.environment} 1"}),
    local.common_tags)
  subnet_id = var.public_subnets_ids[0]
    metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }
}



resource "aws_route53_record" "ec2_voice_api_alb_cname_record" {
  depends_on = [aws_instance.instance1]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "voice-api-${var.environment}"
  records = [module.alb.lb_dns_name]
  ttl     = 300
}


resource "aws_route53_record" "instance1" {
  depends_on = [aws_instance.instance1]
  zone_id = var.openline_hosted_zone_id
  type    = "A"
  name    = "voice-api-1.voice-api-${var.environment}"
  records = [aws_instance.instance1.private_ip]
  ttl     = 300
}

