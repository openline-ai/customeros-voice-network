#!/bin/bash
terraform init -reconfigure -backend-config="key=voice-network/ec2-asterisk"
terraform validate
terraform apply -var-file="../../../variables.tfvars"
