package services

import (
	"agi/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type WebhookClient interface {
	StartCallEvent()
	AnwswerCallEvent()
	EndCallEvent(fromCaller bool)
	UploadFile(filename string) error
}

type WebhookClientImpl struct {
	correlationId    string
	recordingWebhook *string
	callEventWebhook *string
	apiKey           *string
	startTime        *time.Time
	answeredTime     *time.Time
	from             *model.CallEventParty
	to               *model.CallEventParty
}

type CallEvent struct {
	Version       string                `json:"version,default=1.0"`
	CorrelationId string                `json:"correlation_id"`
	Event         string                `json:"event"`
	From          *model.CallEventParty `json:"from"`
	To            *model.CallEventParty `json:"to"`
}

type CallEventStart struct {
	CallEvent
	StartTime time.Time `json:"start_time"`
}

type CallEventAnswered struct {
	CallEvent
	StartTime    time.Time `json:"start_time"`
	AnsweredTime time.Time `json:"answered_time"`
}

type CallEventEnd struct {
	CallEvent
	StartTime    *time.Time `json:"start_time,omitempty"`
	AnsweredTime *time.Time `json:"answered_time,omitempty"`
	EndTime      time.Time  `json:"end_time"`
	Duration     int64      `json:"duration"`
	FromCaller   bool       `json:"from_caller"`
}

func (fsc *WebhookClientImpl) buildBaseCallEvent(event string) CallEvent {
	return CallEvent{
		Version:       "1.0",
		CorrelationId: fsc.correlationId,
		Event:         event,
		From:          fsc.from,
		To:            fsc.to,
	}

}

func (fsc *WebhookClientImpl) StartCallEvent() {
	timeNow := time.Now()
	fsc.startTime = &timeNow
	callEvent := CallEventStart{
		CallEvent: fsc.buildBaseCallEvent("CALL_START"),
		StartTime: *fsc.startTime,
	}
	// create an http json body with the callEvent
	body, err := json.Marshal(callEvent)
	if err != nil {
		log.Printf("StartCallEvent: could not marshal callEvent: %s\n", err)
		return
	}
	fsc.sendCallEvent("StartCallEvent", body)
}

func (fsc *WebhookClientImpl) AnwswerCallEvent() {
	timeNow := time.Now()
	fsc.answeredTime = &timeNow
	callEvent := CallEventAnswered{
		CallEvent:    fsc.buildBaseCallEvent("CALL_ANSWERED"),
		StartTime:    *fsc.startTime,
		AnsweredTime: timeNow,
	}
	// create an http json body with the callEvent
	body, err := json.Marshal(callEvent)
	if err != nil {
		log.Printf("AnswerCallEvent: could not marshal callEvent: %s\n", err)
		return
	}
	fsc.sendCallEvent("AnswerCallEvent", body)
}

func (fsc *WebhookClientImpl) EndCallEvent(fromCaller bool) {
	timeNow := time.Now()
	callEvent := CallEventEnd{
		CallEvent:    fsc.buildBaseCallEvent("CALL_END"),
		StartTime:    fsc.startTime,
		AnsweredTime: fsc.answeredTime,
		EndTime:      timeNow,
		FromCaller:   fromCaller,
	}
	if fsc.answeredTime != nil {
		callEvent.Duration = timeNow.Sub(*fsc.answeredTime).Milliseconds()
	} else {
		callEvent.Duration = 0
	}
	// create an http json body with the callEvent
	body, err := json.Marshal(callEvent)
	if err != nil {
		log.Printf("EndCallEvent: could not marshal callEvent: %s\n", err)
		return
	}
	fsc.sendCallEvent("EndCallEvent", body)
}

func (fsc *WebhookClientImpl) sendCallEvent(event string, body []byte) error {
	if fsc.callEventWebhook == nil {
		return nil
	}
	// create an http request
	req, err := http.NewRequest("POST", *fsc.callEventWebhook, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if fsc.apiKey != nil {
		req.Header.Add("X-API-KEY", *fsc.apiKey)
	}
	log.Printf("sending request to %s: content: %s\n", *fsc.callEventWebhook, string(body))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("%s: could not send request: %s\n", event, err)
		return fmt.Errorf("%s: could not send request: %s\n", event, err)
	}
	defer resp.Body.Close()
	responseMsg, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s: got response: %d\n body: %s", event, resp.StatusCode, string(responseMsg))
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: got response: %d\n", event, resp.StatusCode)
	}
	return nil
}

func (fsc *WebhookClientImpl) UploadFile(filename string) error {
	if fsc.recordingWebhook == nil {
		return nil
	}
	file, _ := os.Open(filename)
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("audio", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.WriteField("correlationId", fsc.correlationId)
	writer.Close()
	r, _ := http.NewRequest("POST", *fsc.recordingWebhook, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	r.Header.Add("Accept", "application/json")
	if fsc.apiKey != nil {
		r.Header.Add("X-API-KEY", *fsc.apiKey)
	}

	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		log.Printf("UploadFile: could not send request: %s\n", err)
		return err
	}

	log.Printf("UploadFile: got response: %d!\n%s", res.StatusCode, res.Body)
	if res.StatusCode != 200 {
		return errors.New("UploadFile: got response: " + string(res.StatusCode))
	}
	return nil
}

func newWebhookClient(channelVar *model.ChannelVar) WebhookClient {
	return &WebhookClientImpl{
		correlationId:    channelVar.Uuid,
		recordingWebhook: channelVar.RecordingEventWebhook,
		callEventWebhook: channelVar.CallEventWebhook,
		apiKey:           channelVar.WebookApiKey,
		from:             channelVar.From,
		to:               channelVar.To,
	}
}
