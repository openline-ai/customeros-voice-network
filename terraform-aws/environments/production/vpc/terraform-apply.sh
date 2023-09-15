#!/bin/bash
terraform init -reconfigure -backend-config="key=voice-network/vpc"
terraform validate
terraform apply -var-file="../variables.tfvars"