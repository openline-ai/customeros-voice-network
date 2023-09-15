package server

import (
	"agi/mocks"
	"agi/model"
	"encoding/json"
	"fmt"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/client/arimocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var mockProfile = map[string]string{
	"call_webhook":      "http://localhost:8080/call",
	"recording_webhook": "http://localhost:8080/recording",
	"webhook_api_key":   "my_awesome_api_key",
}

type setupChannelResult struct {
	dialKey         *ari.Key
	subHangupChan   chan ari.Event
	subAnswerChan   chan ari.Event
	aHangupChan     chan ari.Event
	inMixEventChan  chan ari.Event
	outMixEventChan chan ari.Event
}

func TestAriHandler_ESIM_TO_PSTN(t *testing.T) {
	mockFactory := &mocks.ServiceFactoryMocks{}
	ariHandler := NewRecordAriHandler(mockFactory)

	assert := assert.New(t)

	vars := make(map[string]map[string]string)
	//var varResult string
	key := ari.NewKey(ari.ChannelKey, "exampleChannel")

	vars[key.ID] = make(map[string]string)
	vars[key.ID]["UUID"] = "my_awesome_uuid"
	vars[key.ID]["DEST"] = "sip:+15551234567@example.org"
	vars[key.ID]["KAMAILIO_IP"] = "127.0.0.1"
	vars[key.ID]["ENDPOINT_NAME"] = "pstn_client"
	vars[key.ID]["ORIGIN_CARRIER"] = ""
	vars[key.ID]["DEST_CARRIER"] = "carrier1"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Endpoint-Type)"] = "pstn"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Dest-Endpoint-Type)"] = "pstn"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-User)"] = "sip:AgentSmith@openline.ai"
	vars[key.ID]["PJSIP_HEADER(read,From)"] = "\"Test Mobile 0001\" <sip:test0001@openline.ai>;tag=as7b0a0b0a"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Dest-User)"] = "" // no dest user because it is a PSTN call
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Dest)"] = "sip:+15551234567@example.org"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Voicemail-Info)"] = ""
	vars[key.ID]["CALLERID(num)"] = "AgentSmith"

	profileBytes, _ := json.Marshal(mockProfile)
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Profile-Info)"] = string(profileBytes)

	c := setupClient()

	// iterate the vars map
	channelResult := setupChannel(vars, key, c, mockFactory)

	inBoundChannelHandle := ari.NewChannelHandle(key, c.Channel(), nil)

	mockFactory.WebhookClientMock.On("EndCallEvent", false).Once().Return()

	//we send hangup on the B leg, so we expect ari to hangup the a leg
	c.Channel().(*arimocks.Channel).On("Hangup", key, "normal").Return(nil).Once()

	go func() {
		channelResult.subAnswerChan <- &ari.ChannelStateChange{Channel: ari.ChannelData{State: "Up"}}
		<-time.After(500 * time.Millisecond)
		channelResult.subHangupChan <- &ari.ChannelHangupRequest{}
	}()

	ariHandler.App(c, inBoundChannelHandle)

	channelResult.inMixEventChan <- &ari.BridgeDestroyed{
		EventData: ari.EventData{Type: ari.Events.BridgeDestroyed},
	}
	channelResult.outMixEventChan <- &ari.BridgeDestroyed{
		EventData: ari.EventData{Type: ari.Events.BridgeDestroyed},
	}
	<-time.After(500 * time.Millisecond)

	mock.AssertExpectationsForObjects(t, &mockFactory.WebhookClientMock, &mockFactory.InboundRtpServerMock, &mockFactory.OutboundRtpServerMock, c.Channel(), c.Bridge())
	assert.Equal(vars[channelResult.dialKey.ID]["UUID"], "my_awesome_uuid")
	assert.Equal(vars[channelResult.dialKey.ID]["PJSIP_HEADER(add,X-Openline-UUID)"], "my_awesome_uuid")
	assert.Equal(vars[channelResult.dialKey.ID]["DEST"], "sip:+15551234567@example.org")
	assert.Equal(vars[channelResult.dialKey.ID]["KAMAILIO_IP"], "127.0.0.1")
	assert.Equal(vars[channelResult.dialKey.ID]["PJSIP_HEADER(add,X-Openline-Dest-Carrier)"], "carrier1")
	assert.Equal(vars[channelResult.dialKey.ID]["DEST_CARRIER"], "carrier1")

	assert.NotNil(mockFactory.WebhookClientChanelVars)
	assert.Equal(mockFactory.WebhookClientChanelVars.From.Type, model.CALL_EVENT_TYPE_SIP)
	assert.Equal(mockFactory.WebhookClientChanelVars.To.Type, model.CALL_EVENT_TYPE_PSTN)
	assert.Equal(*mockFactory.WebhookClientChanelVars.From.Mailto, "AgentSmith@openline.ai")
	assert.Equal(*mockFactory.WebhookClientChanelVars.To.Tel, "+15551234567")
	assert.Nil(mockFactory.WebhookClientChanelVars.From.Tel)
	assert.Nil(mockFactory.WebhookClientChanelVars.To.Mailto)

}

func TestAriHandler_PSTN_TO_PSTN(t *testing.T) {
	mockFactory := &mocks.ServiceFactoryMocks{}
	ariHandler := NewRecordAriHandler(mockFactory)

	assert := assert.New(t)

	vars := make(map[string]map[string]string)
	//var varResult string
	key := ari.NewKey(ari.ChannelKey, "exampleChannel")

	vars[key.ID] = make(map[string]string)
	vars[key.ID]["UUID"] = "my_awesome_uuid"
	vars[key.ID]["DEST"] = "sip:+15551234567@example.org"
	vars[key.ID]["KAMAILIO_IP"] = "127.0.0.1"
	vars[key.ID]["ENDPOINT_NAME"] = "pstn"
	vars[key.ID]["ORIGIN_CARRIER"] = "carrier1"
	vars[key.ID]["DEST_CARRIER"] = "carrier1"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Endpoint-Type)"] = "pstn"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Dest-Endpoint-Type)"] = "pstn"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-User)"] = "" // no user - random person on the PSTN
	vars[key.ID]["PJSIP_HEADER(read,From)"] = "<sip:+32555123456@openline.ai>;tag=as7b0a0b0a"
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Dest-User)"] = "sip:AgentSmith@openline.ai" // user originally called
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Dest)"] = "sip:+15551234567@example.org"    // pstn destination to forward to
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Voicemail-Info)"] = ""
	vars[key.ID]["CALLERID(num)"] = "AgentSmith"

	profileBytes, _ := json.Marshal(mockProfile)
	vars[key.ID]["PJSIP_HEADER(read,X-Openline-Profile-Info)"] = string(profileBytes)

	c := setupClient()

	// iterate the vars map
	channelResult := setupChannel(vars, key, c, mockFactory)

	inBoundChannelHandle := ari.NewChannelHandle(key, c.Channel(), nil)

	mockFactory.WebhookClientMock.On("EndCallEvent", false).Once().Return()

	//we send hangup on the B leg, so we expect ari to hangup the a leg
	c.Channel().(*arimocks.Channel).On("Hangup", key, "normal").Return(nil).Once()

	go func() {
		channelResult.subAnswerChan <- &ari.ChannelStateChange{Channel: ari.ChannelData{State: "Up"}}
		<-time.After(500 * time.Millisecond)
		channelResult.subHangupChan <- &ari.ChannelHangupRequest{}
	}()

	ariHandler.App(c, inBoundChannelHandle)

	channelResult.inMixEventChan <- &ari.BridgeDestroyed{
		EventData: ari.EventData{Type: ari.Events.BridgeDestroyed},
	}
	channelResult.outMixEventChan <- &ari.BridgeDestroyed{
		EventData: ari.EventData{Type: ari.Events.BridgeDestroyed},
	}
	<-time.After(500 * time.Millisecond)

	mock.AssertExpectationsForObjects(t, &mockFactory.WebhookClientMock, &mockFactory.InboundRtpServerMock, &mockFactory.OutboundRtpServerMock, c.Channel(), c.Bridge())
	assert.Equal(vars[channelResult.dialKey.ID]["UUID"], "my_awesome_uuid")
	assert.Equal(vars[channelResult.dialKey.ID]["PJSIP_HEADER(add,X-Openline-UUID)"], "my_awesome_uuid")
	assert.Equal(vars[channelResult.dialKey.ID]["DEST"], "sip:+15551234567@example.org")
	assert.Equal(vars[channelResult.dialKey.ID]["KAMAILIO_IP"], "127.0.0.1")
	assert.Equal(vars[channelResult.dialKey.ID]["PJSIP_HEADER(add,X-Openline-Dest-Carrier)"], "carrier1")
	assert.Equal(vars[channelResult.dialKey.ID]["DEST_CARRIER"], "carrier1")
	assert.Equal(vars[channelResult.dialKey.ID]["PJSIP_HEADER(add,X-Openline-Origin-Carrier)"], "carrier1")
	assert.Equal(vars[channelResult.dialKey.ID]["ORIGIN_CARRIER"], "carrier1")

	assert.NotNil(mockFactory.WebhookClientChanelVars)
	assert.Equal(mockFactory.WebhookClientChanelVars.From.Type, model.CALL_EVENT_TYPE_PSTN)
	assert.Equal(mockFactory.WebhookClientChanelVars.To.Type, model.CALL_EVENT_TYPE_PSTN)
	assert.Equal(*mockFactory.WebhookClientChanelVars.From.Tel, "+32555123456")
	assert.Equal(*mockFactory.WebhookClientChanelVars.To.Tel, "+15551234567")
	assert.Equal(*mockFactory.WebhookClientChanelVars.To.Mailto, "AgentSmith@openline.ai")
	assert.Nil(mockFactory.WebhookClientChanelVars.From.Mailto)

}

func setupChannel(vars map[string]map[string]string, key *ari.Key, client *arimocks.Client, factory *mocks.ServiceFactoryMocks) setupChannelResult {
	factory.WebhookClientMock.On("StartCallEvent").Once().Return()
	factory.WebhookClientMock.On("AnwswerCallEvent").Once().Return()

	factory.InboundRtpServerMock.On("Address").Return("127.0.0.1:8000")
	factory.OutboundRtpServerMock.On("Address").Return("127.0.0.1:9000")

	factory.InboundRtpServerMock.On("Listen").Return(nil).Once()
	factory.OutboundRtpServerMock.On("Listen").Return(nil).Once()

	factory.InboundRtpServerMock.On("Close").Return().Once()
	factory.OutboundRtpServerMock.On("Close").Return().Once()

	channel := client.Channel().(*arimocks.Channel)
	bridge := client.Bridge().(*arimocks.Bridge)
	for k, v := range vars[key.ID] {
		if v != "" {
			channel.On("GetVariable", key, k).Return(v, nil)
		} else {
			channel.On("GetVariable", key, k).Return("", fmt.Errorf("Variable not found"))
		}
	}
	dialedKey := ari.NewKey(ari.ChannelKey, "managed-dialed-channel-exampleChannel")
	outboundChannelHandle := ari.NewChannelHandle(dialedKey, channel, nil)

	vars[dialedKey.ID] = make(map[string]string)
	channel.On("SetVariable", dialedKey, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Run(func(args mock.Arguments) {
		vars[dialedKey.ID][args.Get(1).(string)] = args.Get(2).(string)
	})

	//var eventWakeup = make(chan bool)
	subAnswer := &arimocks.Subscription{}
	var subAnswerChannel = make(chan ari.Event)
	channel.On("Subscribe", dialedKey, ari.Events.ChannelStateChange).Return(subAnswer)
	subAnswer.On("Events").Return((<-chan ari.Event)(subAnswerChannel))
	subHangup := &arimocks.Subscription{}
	var subHangupChannel = make(chan ari.Event)
	channel.On("Subscribe", dialedKey, ari.Events.ChannelHangupRequest).Return(subHangup)
	subHangup.On("Events").Return((<-chan ari.Event)(subHangupChannel))

	aHangup := &arimocks.Subscription{}
	var aHangupChannel = make(chan ari.Event)
	channel.On("Subscribe", key, ari.Events.ChannelHangupRequest).Return(aHangup)
	aHangup.On("Events").Return((<-chan ari.Event)(aHangupChannel))
	channel.On("Create", key, mock.AnythingOfType("ari.ChannelCreateRequest")).Return(outboundChannelHandle, nil)
	channel.On("Dial", dialedKey, vars[key.ID]["CALLERID(num)"], mock.AnythingOfType("time.Duration")).Return(nil)

	bridgeKey := ari.NewKey(ari.BridgeKey, uuid.New().String())
	bridgeHandle := ari.NewBridgeHandle(bridgeKey, bridge, nil)
	bridge.On("Create", mock.AnythingOfType("*ari.Key"), "mixing", "managed-dialBridge-"+key.ID).Return(bridgeHandle, nil)
	bridge.On("AddChannel", bridgeKey, outboundChannelHandle.ID()).Return(nil).Once()
	bridge.On("AddChannel", bridgeKey, key.ID).Return(nil).Once()
	bridge.On("Delete", bridgeKey).Return(nil).Once()

	inboundSnoopKey := ari.NewKey(ari.ChannelKey, "managed-in-snoop-"+key.ID)
	inboundSnoopChannelHandle := ari.NewChannelHandle(inboundSnoopKey, channel, nil)
	outboundSnoopKey := ari.NewKey(ari.ChannelKey, "managed-out-snoop-"+key.ID)
	outboundSnoopChannelHandle := ari.NewChannelHandle(outboundSnoopKey, channel, nil)

	channel.On("Snoop", key, mock.AnythingOfType("string"), &ari.SnoopOptions{App: client.ApplicationName(), Spy: ari.DirectionIn}).Return(inboundSnoopChannelHandle, nil).Once()
	channel.On("Snoop", key, mock.AnythingOfType("string"), &ari.SnoopOptions{App: client.ApplicationName(), Spy: ari.DirectionOut}).Return(outboundSnoopChannelHandle, nil).Once()

	keyExternalMediaIn := ari.NewKey(ari.ChannelKey, "managed-in-"+key.ID)
	channel.On("ExternalMedia", (*ari.Key)(nil), ari.ExternalMediaOptions{
		App:          client.ApplicationName(),
		ExternalHost: factory.InboundRtpServerMock.Address(),
		Format:       "slin48",
		ChannelID:    keyExternalMediaIn.ID,
	}).Return(ari.NewChannelHandle(keyExternalMediaIn, channel, nil), nil).Once()
	channel.On("Hangup", keyExternalMediaIn, "").Return(nil).Once()

	keyExternalMediaOut := ari.NewKey(ari.ChannelKey, "managed-out-"+key.ID)
	channel.On("ExternalMedia", (*ari.Key)(nil), ari.ExternalMediaOptions{
		App:          client.ApplicationName(),
		ExternalHost: factory.OutboundRtpServerMock.Address(),
		Format:       "slin48",
		ChannelID:    keyExternalMediaOut.ID,
	}).Return(ari.NewChannelHandle(keyExternalMediaOut, channel, nil), nil).Once()
	channel.On("Hangup", keyExternalMediaOut, "").Return(nil).Once()

	inMixBridgeHandle := ari.NewBridgeHandle(ari.NewKey(ari.ChannelKey, keyExternalMediaIn.ID), bridge, nil)
	outMixBridgeHandle := ari.NewBridgeHandle(ari.NewKey(ari.ChannelKey, keyExternalMediaOut.ID), bridge, nil)

	bridge.On("Create", mock.AnythingOfType("*ari.Key"), "mixing", inMixBridgeHandle.ID()).Return(inMixBridgeHandle, nil)
	bridge.On("Create", mock.AnythingOfType("*ari.Key"), "mixing", outMixBridgeHandle.ID()).Return(outMixBridgeHandle, nil)

	bridge.On("AddChannel", inMixBridgeHandle.Key(), keyExternalMediaIn.ID).Return(nil).Once()
	bridge.On("AddChannel", inMixBridgeHandle.Key(), inboundSnoopKey.ID).Return(nil).Once()
	bridge.On("Delete", inMixBridgeHandle.Key()).Return(nil).Once()
	bridge.On("AddChannel", outMixBridgeHandle.Key(), keyExternalMediaOut.ID).Return(nil).Once()
	bridge.On("AddChannel", outMixBridgeHandle.Key(), outboundSnoopKey.ID).Return(nil).Once()
	bridge.On("Delete", outMixBridgeHandle.Key()).Return(nil).Once()

	inBridgeSubscription := &arimocks.Subscription{}
	bridge.On("Subscribe", inMixBridgeHandle.Key(), mock.Anything, mock.Anything, mock.Anything).Return(inBridgeSubscription, nil)

	outBridgeSubscription := &arimocks.Subscription{}
	bridge.On("Subscribe", outMixBridgeHandle.Key(), mock.Anything, mock.Anything, mock.Anything).Return(outBridgeSubscription, nil)

	// don't care about this event
	bridge.On("Data", mock.AnythingOfType("*ari.Key")).Return(&ari.BridgeData{}, nil)

	var inMixEventChannel = make(chan ari.Event)
	inBridgeSubscription.On("Events").Return((<-chan ari.Event)(inMixEventChannel))
	inBridgeSubscription.On("Cancel").Return()

	var outMixEventChannel = make(chan ari.Event)
	outBridgeSubscription.On("Events").Return((<-chan ari.Event)(outMixEventChannel))
	outBridgeSubscription.On("Cancel").Return()

	return setupChannelResult{
		dialKey:         dialedKey,
		subHangupChan:   subHangupChannel,
		subAnswerChan:   subAnswerChannel,
		aHangupChan:     aHangupChannel,
		inMixEventChan:  inMixEventChannel,
		outMixEventChan: outMixEventChannel,
	}
}

func setupClient() *arimocks.Client {
	c := &arimocks.Client{}
	inboundChannel := &arimocks.Channel{}
	myBridge := &arimocks.Bridge{}

	c.On("Channel").Return(inboundChannel)
	c.On("ApplicationName").Return("test-application")
	c.On("Bridge").Return(myBridge)

	return c
}
