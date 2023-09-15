provider "aws" {
  region = var.aws_region
}

locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
    CostIdentifier = "Voice Network"
    Service = "Asterisk"
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
    values = ["asterisk-server-ami_${var.environment}"]
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

    from_port = 5060
    to_port   = 5060
    protocol  = "udp"
  }
  ingress {
    cidr_blocks = ["0.0.0.0/0"]

    from_port = 10000
    to_port   = 20000
    protocol  = "udp"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    tomap({ CostIdentifier = "EC2 Asterisk" }),
    local.common_tags)
}

module "cloudwatch_iam_profile" {
  source = "../../../iam_role/push_logs_to_cloudwatch"
  environment = var.environment
  service = "asterisk"

}

module "s3_voicemail_bucket" {
  source = "../../../iam_role/s3_bucket"
  environment = var.environment
  service = "asterisk"
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
  name = "asterisk-access-role"
  assume_role_policy = data.aws_iam_policy_document.role_policy.json
}


resource "aws_iam_instance_profile" "instance_profile" {
  name = "${var.environment}-asterisk-access-profile"
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
  instance_type   = var.asterisk_instance_type
  vpc_security_group_ids = [aws_security_group.sg.id]
  key_name = var.ec2_ssh_key_name
  iam_instance_profile =  aws_iam_instance_profile.instance_profile.name
  associate_public_ip_address = true

  tags = merge(
    tomap({Name = "Asterisk Server ${var.environment} 1"}),
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
  instance_type   = var.asterisk_instance_type
  vpc_security_group_ids = [aws_security_group.sg.id]
  key_name = var.ec2_ssh_key_name
  iam_instance_profile =  aws_iam_instance_profile.instance_profile.name
  associate_public_ip_address = true

  tags = merge(
    tomap({Name = "Asterisk Server ${var.environment} 2"}),
    local.common_tags)
  subnet_id = var.public_subnets_ids[0]
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }
}

resource "aws_route53_record" "ast_instance1_cname" {
  depends_on = [aws_instance.instance1]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "asterisk-${var.environment}-1"
  records = [aws_instance.instance1.private_dns]
  ttl     = 300
}

resource "aws_route53_record" "ast_instance2_cname" {
  depends_on = [aws_instance.instance2]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "asterisk-${var.environment}-2"
  records = [aws_instance.instance2.private_dns]
  ttl     = 300
}