#!/bin/bash
terraform init -reconfigure -backend-config="key=voice-network/ec2-secrets"
terraform validate
terraform apply -var-file="../../variables.tfvars"
