package mocks

import (
	"github.com/stretchr/testify/mock"
	"log"
	"time"
)

type WebhookClientMock struct {
	mock.Mock
	startTime    *time.Time
	answeredTime *time.Time
}

func (fsc *WebhookClientMock) StartCallEvent() {
	timeNow := time.Now()
	fsc.startTime = &timeNow
	fsc.Called()
}

func (fsc *WebhookClientMock) AnwswerCallEvent() {
	timeNow := time.Now()
	fsc.answeredTime = &timeNow
	log.Printf("StartCallEvent: %v\n", *fsc.startTime)
	fsc.Called()
}

func (fsc *WebhookClientMock) EndCallEvent(fromCaller bool) {
	fsc.Called(fromCaller)
}

func (fsc *WebhookClientMock) UploadFile(filename string) error {
	args := fsc.Called(filename)
	return args.Error(0)
}
