#!/bin/bash
terraform init -reconfigure -backend-config="key=voice-network/rds-postgres"
terraform validate
terraform apply -var-file="../variables.tfvars"