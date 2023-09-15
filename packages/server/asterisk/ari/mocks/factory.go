package mocks

import (
	"agi/model"
	"agi/services"
	"gopkg.in/ini.v1"
)

type ServiceFactoryMocks struct {
	InboundRtpServerMock    RtpServerMock
	OutboundRtpServerMock   RtpServerMock
	WebhookClientMock       WebhookClientMock
	WebhookClientChanelVars *model.ChannelVar
}

func (sfm *ServiceFactoryMocks) NewRtpServer(cd *model.CallMetadata) services.RtpServer {
	if cd.Direction == model.IN {
		return sfm.InboundRtpServerMock
	}
	return sfm.OutboundRtpServerMock
}

func (sfm *ServiceFactoryMocks) NewWebhookClient(channelVar *model.ChannelVar) services.WebhookClient {
	sfm.WebhookClientChanelVars = channelVar
	return &sfm.WebhookClientMock
}

func (sfm *ServiceFactoryMocks) NewS3Client(cfg *ini.File) services.S3Client {
	return nil
}
