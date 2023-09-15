#!/bin/bash

mkdir /etc/voice-api
echo "VOICE_API_SERVER_ADDRESS=:8080" > /etc/voice-api/env
echo "VOICE_API_KEY=$VOICE_API_KEY" >> /etc/voice-api/env
echo "VOICE_API_HOST=$VOICE_API_HOST" >> /etc/voice-api/env
echo "DB_HOST=$DB_HOST" >> /etc/voice-api/env
echo "DB_PORT=5432" >> /etc/voice-api/env
echo "DB_NAME=$DB_NAME" >> /etc/voice-api/env
echo "DB_USER=$DB_USER" >> /etc/voice-api/env
echo "DB_PASSWORD=$DB_PASSWORD" >> /etc/voice-api/env
echo "GIN_MODE=release" >> /etc/voice-api/env
echo "AWS_S3_REGION=$AWS_S3_REGION" >> /etc/voice-api/env
echo "AWS_S3_BUCKET=$AWS_S3_BUCKET" >> /etc/voice-api/env

