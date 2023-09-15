#!/bin/bash

declare -a environment=voice-network
declare -a prefix=terraform-state-lock

declare -a mandatory=(vpc rds-postgres)

declare -a ec2_services=(ec2-secrets-voice-api ec2-secrets-kamailio ec2-asterisk ec2-homer ec2-kamailio ec2-voice-api)
declare -a monitoring=(alarms alarms-lambdas eks-alarms)
tables=("${mandatory[@]}" "${ec2_services[@]}" "${monitoring[@]}")

for i in "${tables[@]}"
do
   declare tableName=${prefix}-${environment}-$i

   aws dynamodb create-table \
       --table-name ${tableName} \
       --region eu-west-1 \
       --attribute-definitions \
           AttributeName=LockID,AttributeType=S \
       --key-schema \
           AttributeName=LockID,KeyType=HASH \
       --billing-mode PAY_PER_REQUEST \
       --tags Key=Environment,Value=${environment} Key=CreatedBy,Value=GenerateDynamoDbScriptForTerraforms Key=CostIdentifier,Value=DynamoDbTableForTerraformLock \
       --table-class STANDARD > /tmp/null &
done

wait
