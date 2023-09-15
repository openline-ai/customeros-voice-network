variable "region" {
	type=string
	default="eu-west-2"
	sensitive=false
}

variable "environment" {
	type=string
	default="openline-network"
	sensitive=false
}

data "amazon-parameterstore" "db_host" {
  name = "/config/voice-api_${var.environment}/db_host"
  with_decryption = true
}

data "amazon-parameterstore" "db_user" {
  name = "/config/voice-api_${var.environment}/db_user"
  with_decryption = true
}

data "amazon-parameterstore" "db_database" {
  name = "/config/voice-api_${var.environment}/db_database"
  with_decryption = true
}

data "amazon-parameterstore" "db_password" {
  name = "/config/voice-api_${var.environment}/db_password"
  with_decryption = true
}
data "amazon-parameterstore" "voice_api_key" {
  name = "/config/voice-api_${var.environment}/key"
  with_decryption = true
}

data "amazon-parameterstore" "voice_api_hostname" {
  name = "/config/voice-api_${var.environment}/hostname"
  with_decryption = true
}

data "amazon-parameterstore" "aws_s3_bucket" {
  name = "/config/voice-api_${var.environment}/aws_s3_bucket"
  with_decryption = true
}

data "amazon-parameterstore" "aws_s3_region" {
  name = "/config/voice-api_${var.environment}/aws_s3_region"
  with_decryption = true
}

# usage example of the data source output
locals {
  db_user   = data.amazon-parameterstore.db_user.value
  db_database   = data.amazon-parameterstore.db_database.value
  db_host   = data.amazon-parameterstore.db_host.value
  db_password   = data.amazon-parameterstore.db_password.value
  voice_api_key   = data.amazon-parameterstore.voice_api_key.value
  voice_api_hostname   = data.amazon-parameterstore.voice_api_hostname.value
  aws_s3_bucket =  data.amazon-parameterstore.aws_s3_bucket.value
  aws_s3_region =  data.amazon-parameterstore.aws_s3_region.value
}

packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
  }
}


source "amazon-ebs" "ubuntu" {
  ami_name      = "voice-api-server-ami_${var.environment}"
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
  name    = "build-voice-api-image"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]
  provisioner "shell" {
    inline = [
      "sudo sh -c 'add-apt-repository universe && apt-get update'",
      "sudo sh -c 'apt-get install -y python2 golang sox libsox-fmt-all'",
      "mkdir /tmp/voice-api/",

    ]
  }
  provisioner "file" { 
	source = "routes"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "schema"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "config"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "go.mod"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "go.sum"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "main.go"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "scripts"
	destination = "/tmp/voice-api/"
  }
  provisioner "file" { 
	source = "awslogs"
	destination = "/tmp/voice-api/"
  }
  provisioner "shell" {
    inline = [
      "sudo sh -c 'cd /tmp/voice-api;go install github.com/swaggo/swag/cmd/swag@latest;go mod download;PATH=$PATH:$HOME/go/bin go generate;go build -o /usr/local/bin/voice-api'",
      "sudo sh -c 'cd /tmp/; curl https://s3.amazonaws.com/aws-cloudwatch/downloads/latest/awslogs-agent-setup.py -O; chmod a+x awslogs-agent-setup.py'",
      "sudo sh -c 'cd /tmp/; python2 ./awslogs-agent-setup.py -r ${var.region} -n -c /tmp/voice-api/awslogs/awslogs.conf'",
      "sudo sh -c 'chmod a+x /tmp/voice-api/scripts/genconf.sh;DB_HOST=\"${local.db_host}\" DB_USER=\"${local.db_user}\" DB_PASSWORD=\"${local.db_password}\" DB_NAME=\"${local.db_database}\" VOICE_API_KEY=\"${local.voice_api_key}\" VOICE_API_HOST=\"${local.voice_api_hostname}\" AWS_S3_REGION=\"${local.aws_s3_region}\" AWS_S3_BUCKET=\"${local.aws_s3_bucket}\" /tmp/voice-api/scripts/genconf.sh'",
      "sudo sh -c 'mv /tmp/voice-api/scripts/voice_api.service /etc/systemd/system'",
      "sudo sh -c 'chmod 644 /etc/systemd/system/voice_api.service'",
      "sudo sh -c 'systemctl enable voice_api.service'",
    ]
  }
}

