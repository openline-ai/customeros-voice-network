#!/bin/bash
terraform init -reconfigure -backend-config="key=voice-network/ec2-homer"
terraform validate
terraform apply -var-file="../../../variables.tfvars"
