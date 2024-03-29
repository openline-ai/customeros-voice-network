variable "region" {
	type=string
	default="eu-west-2"
	sensitive=false
}

variable "environment" {
	type=string
	default="openline-dev"
	sensitive=false
}

packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

locals {
  opus_codec  = "asterisk-18.0/x86-64/codec_opus-18.0_current-x86_64"
  aws_bucket  = "${var.environment}-voicemail-bucket"
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "asterisk-server-ami_${var.environment}"
  instance_type = "t2.micro"
  region        = "${var.region}"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
}

build {
  name    = "build-asterisk-image"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]
  provisioner "shell" {
    inline = [
      "sudo sh -c 'add-apt-repository universe && apt-get update'",
      "sudo sh -c 'apt-get install -y asterisk asterisk-core-sounds-en-wav sox golang collectd'",
      "sudo sh -c 'mkdir -p /usr/src/codecs/opus'",
      "sudo sh -c 'cd /usr/src/codecs/opus && curl -sL http://downloads.digium.com/pub/telephony/codec_opus/${local.opus_codec}.tar.gz | tar --strip-components 1 -xz'",
      "sudo sh -c 'cp /usr/src/codecs/opus/*.so /usr/lib/x86_64-linux-gnu/asterisk/modules/'",
      "sudo sh -c 'cp /usr/src/codecs/opus/codec_opus_config-en_US.xml /usr/share/asterisk/documentation/'",
      "sudo sh -c 'rm /usr/lib/x86_64-linux-gnu/asterisk/modules/format_ogg_opus_open_source.so'",
      "mkdir /tmp/asterisk/",

    ]
  }
  provisioner "file" { 
	source = "conf"
	destination = "/tmp/asterisk/"
  }
  provisioner "file" { 
	source = "scripts"
	destination = "/tmp/asterisk/"
  }
  provisioner "file" { 
	source = "ari"
	destination = "/tmp/asterisk/"
  }
  provisioner "file" { 
	source = "awslogs"
	destination = "/tmp/asterisk/"
  }
  provisioner "shell" {
    inline = [
      "sudo sh -c 'cp -v /tmp/asterisk/conf/* /etc/asterisk/'",
      "sudo sh -c 'cp -v /tmp/asterisk/scripts/asterisk_network_setup.sh /usr/sbin/'",
      "sudo sh -c 'chmod a+x /tmp/asterisk/scripts/asterisk_network_setup.sh'",
      "sudo sh -c 'mv /tmp/asterisk/scripts/asterisk.service /etc/systemd/system'",
      "sudo sh -c 'chmod 644 /etc/systemd/system/asterisk.service'",
      "sudo sh -c 'mv /tmp/asterisk/scripts/asterisk_ari.service /etc/systemd/system'",
      "sudo sh -c 'chmod 644 /etc/systemd/system/asterisk_ari.service'",
      "sudo sh -c 'systemctl enable asterisk_ari.service'",
      "sudo sh -c 'cd /tmp/; curl https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb -O; dpkg -i amazon-cloudwatch-agent.deb'",
      "sudo sh -c 'cp /tmp/asterisk/awslogs/amazon-cloudwatch-agent.json /opt/aws/amazon-cloudwatch-agent/etc/'",
      "sudo sh -c 'mkdir /var/spool/asterisk/recording;chown asterisk:asterisk /var/spool/asterisk/recording'",
      "sudo sh -c 'sed \"s/!AWS_BUCKET!/${local.aws_bucket}/g\" /etc/asterisk/aws.conf.template| sed \"s/!AWS_REGION!/${var.region}/g\" > /etc/asterisk/aws.conf'",
      "sudo sh -c 'systemctl enable amazon-cloudwatch-agent.service'",
      "sudo sh -c 'cd /tmp/asterisk/ari;go mod download;go build -o /usr/local/bin/record_agi'",
    ]
  }
}

