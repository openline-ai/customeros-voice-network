package services

import (
	"agi/model"
	"gopkg.in/ini.v1"
)

type ServiceFactory interface {
	NewRtpServer(cd *model.CallMetadata) RtpServer
	NewWebhookClient(channelVar *model.ChannelVar) WebhookClient
	NewS3Client(cfg *ini.File) S3Client
}

type ServiceFactoryImpl struct{}

func (ServiceFactoryImpl) NewRtpServer(cd *model.CallMetadata) RtpServer {
	return newRtpServer(cd)
}

func (ServiceFactoryImpl) NewWebhookClient(channelVar *model.ChannelVar) WebhookClient {
	return newWebhookClient(channelVar)
}

func (ServiceFactoryImpl) NewS3Client(cfg *ini.File) S3Client {
	return newS3Client(cfg)
}
