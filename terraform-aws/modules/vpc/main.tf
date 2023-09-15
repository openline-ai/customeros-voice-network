locals {
  common_tags = {
    Environment = var.environment
    CreatedBy   = "Terraform"
    CostIdentifier = "VPC"
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = var.environment
  cidr = var.cidr

  azs             = var.azs
  private_subnets = var.private_subnets
  public_subnets  = var.public_subnets

  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  enable_flow_log                      = true
  create_flow_log_cloudwatch_iam_role  = true
  create_flow_log_cloudwatch_log_group = true

  public_subnet_tags = {
    "kubernetes.io/cluster/${var.environment}" = "shared"
    "kubernetes.io/role/elb"              = 1
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${var.environment}" = "shared"
    "kubernetes.io/role/internal-elb"     = 1
  }

  tags = local.common_tags
}

resource "tls_private_key" "pk" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "kp" {
  key_name   = var.ssh_key_name
  public_key = "${trim(tls_private_key.pk.public_key_openssh, "\n")} ${var.ssh_key_name}"

  tags = local.common_tags

  provisioner "local-exec" {
    command = "echo '${tls_private_key.pk.private_key_pem}' > ./private-key.pem"
  }

  provisioner "local-exec" {
    command = "echo '${trim(tls_private_key.pk.public_key_openssh, "\n")} ${var.ssh_key_name}' > ./public-key.pem"
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "allow_ssh" {
  name = "allow-ssh"
  vpc_id = module.vpc.vpc_id
  dynamic "ingress" {
    for_each = var.ssh_jump_allow_lists
    content {
      cidr_blocks = ingress.value.subnets
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      description = ingress.value.description
    }
  }
  // Terraform removes the default rule
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
  depends_on = [module.vpc]
}

resource "aws_instance" "web" {
  depends_on = [module.vpc, aws_key_pair.kp]
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  key_name = var.ssh_key_name

  associate_public_ip_address = true

  subnet_id              = module.vpc.public_subnets[0]
  vpc_security_group_ids = [aws_security_group.allow_ssh.id]

  tags = merge(
    tomap({Name = "Jump ${var.environment}", Service = "Jump"}),
    local.common_tags)
  
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }  
}

resource "aws_route53_record" "web_cname" {
  depends_on = [aws_instance.web]
  zone_id = var.openline_hosted_zone_id
  type    = "CNAME"
  name    = "jump-${var.environment}"
  records = [aws_instance.web.public_dns]
  ttl     = 300
}
