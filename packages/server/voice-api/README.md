# Set up Development Environment

```
go install github.com/swaggo/swag/cmd/swag@latest
go generate
go build -o bin/voice-plugin
```

## Environment Variables

| param                    | meaning                                              |
|--------------------------|------------------------------------------------------|
| VOICE_API_SERVER_ADDRESS | port for the voice rest api, should be set to :11010 |
| VOICE_API_KEY            | api key for the voice rest api                       |
| VOICE_API_BASE_URL       | swagger host, should be localhost:11010              |
| DB_HOST                  | hostname of postgres db                              |
| DB_PORT                  | port of postgres db                                  |
| DB_NAME                  | database name                                        |
| DB_USER                  | user to log into db as                               |
| DB_PASSWORD              | the database password                                |
| AWS_S3_REGION            | the aws region                                       |
| AWS_S3_BUCKET            | the aws bucket name                                  |


## Accessing the swagger interface
http://localhost:11010/swagger/index.html

you can build an image as follows

```
export AWS_REGION=eu-west-1
packer init aws-ubuntu.pkr.hcl
packer validate aws-ubuntu.pkr.hcl
packer build -var 'region=eu-west-1' -var 'environment=openline-network' aws-ubuntu.pkr.hcl
```
