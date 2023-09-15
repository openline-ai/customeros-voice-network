package server

import (
	"agi/model"
	"agi/services"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/bridgemon"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
)

type RecordAriHandler struct {
	serviceFactory services.ServiceFactory
}

func NewRecordAriHandler(serviceFactory services.ServiceFactory) *RecordAriHandler {
	return &RecordAriHandler{
		serviceFactory: serviceFactory,
	}
}

func setDialVariables(h *ari.ChannelHandle, channelVars *model.ChannelVar) {
	h.SetVariable("PJSIP_HEADER(add,X-Openline-UUID)", channelVars.Uuid)
	h.SetVariable("UUID", channelVars.Uuid)
	h.SetVariable("PJSIP_HEADER(add,X-Openline-DEST)", channelVars.Dest)
	h.SetVariable("DEST", channelVars.Dest)
	h.SetVariable("KAMAILIO_IP", channelVars.KamailioIP)
	if channelVars.OriginCarrier != nil {
		h.SetVariable("PJSIP_HEADER(add,X-Openline-Origin-Carrier)", *channelVars.OriginCarrier)
		h.SetVariable("ORIGIN_CARRIER", *channelVars.OriginCarrier)
	}
	if channelVars.DestCarrier != nil {
		h.SetVariable("PJSIP_HEADER(add,X-Openline-Dest-Carrier)", *channelVars.DestCarrier)
		h.SetVariable("DEST_CARRIER", *channelVars.DestCarrier)
	}
	h.SetVariable("TRANSFER_CONTEXT", "transfer")

}
func (ah *RecordAriHandler) App(cl ari.Client, h *ari.ChannelHandle) {
	log.Printf("Running Record App channel: %s", h.Key().ID)
	channelVars, err := model.GetChannelVars(h, true)
	if err != nil {
		log.Printf("Error getting channel vars: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		h.Busy()
		return
	}

	dialString := "PJSIP/" + channelVars.EndpointName + "/sip:" + channelVars.KamailioIP
	dialedChannel, err := h.Create(ari.ChannelCreateRequest{
		Endpoint:  dialString,
		App:       cl.ApplicationName(),
		ChannelID: "managed-dialed-channel-" + h.ID(),
	})

	if err != nil {
		log.Printf("Error creating outbound channel: %v", err)
		h.Busy()
		return
	}
	setDialVariables(dialedChannel, channelVars)
	subAnswer := dialedChannel.Subscribe(ari.Events.ChannelStateChange)
	subHangup := dialedChannel.Subscribe(ari.Events.ChannelHangupRequest)
	aHangup := h.Subscribe(ari.Events.ChannelHangupRequest)
	id, _ := h.GetVariable("CALLERID(num)")

	dialBridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "managed-dialBridge-"+h.ID())

	if err != nil {
		log.Printf("Error creating bridge: %v", err)
		h.Busy()
		return
	}
	err = dialBridge.AddChannel(h.ID())
	if err != nil {
		log.Printf("Error adding calling channel to bridge: %v", err)
		h.Busy()
		return
	}
	err = dialBridge.AddChannel(dialedChannel.ID())
	if err != nil {
		log.Printf("Error adding dialed channel to bridge: %v", err)
		h.Busy()
		return
	}

	webhookClient := ah.serviceFactory.NewWebhookClient(channelVars)

	webhookClient.StartCallEvent()

	err = cl.Channel().Dial(dialedChannel.Key(), id, 120)
	if err != nil {
		log.Printf("Error dialing: %v", err)
		h.Busy()
		return
	}
	for {
		select {
		case e := <-subAnswer.Events():
			v := e.(*ari.ChannelStateChange)
			log.Printf("Got Channel State Change for channel: %s new state: %s", v.Channel.ID, v.Channel.State)
			if v.Channel.State == "Up" {
				webhookClient.AnwswerCallEvent()
				var counter int = 0
				ah.record(cl, h, model.MakeMetaData(model.IN, channelVars), &counter, webhookClient)
				ah.record(cl, h, model.MakeMetaData(model.OUT, channelVars), &counter, webhookClient)
			}

		case e := <-subHangup.Events():
			webhookClient.EndCallEvent(false)
			v := e.(*ari.ChannelHangupRequest)
			log.Printf("Got Channel Hangup for channel: %s", v.Channel.ID)
			h.Hangup()
			dialBridge.Delete()
			return
		case e := <-aHangup.Events():
			webhookClient.EndCallEvent(true)
			v := e.(*ari.ChannelHangupRequest)
			log.Printf("Got Channel Hangup for channel: %s", v.Channel.ID)
			dialedChannel.Hangup()
			dialBridge.Delete()
			return
		}
	}

}

func (ah *RecordAriHandler) record(cl ari.Client, h *ari.ChannelHandle, metadata *model.CallMetadata, counter *int, webhook services.WebhookClient) {

	snoopOptions := &ari.SnoopOptions{
		App: cl.ApplicationName(),
	}
	if metadata.Direction == model.IN {
		snoopOptions.Spy = ari.DirectionIn
	} else {
		snoopOptions.Spy = ari.DirectionOut
	}
	snoopChannel, err := cl.Channel().Snoop(h.Key(), "managed-"+string(metadata.Direction)+"-snoop-"+h.ID(), snoopOptions)
	if err != nil {
		log.Printf("Error making %s Snoop: %v", metadata.Direction, err)
		return
	}

	rtpServer := ah.serviceFactory.NewRtpServer(metadata)

	log.Printf("%s RTP Server created: %s", metadata.Direction, rtpServer.Address())
	go rtpServer.Listen()
	//go rtpServer.ListenForText()
	mediaChannel, err := cl.Channel().ExternalMedia(nil, ari.ExternalMediaOptions{
		App:          cl.ApplicationName(),
		ExternalHost: rtpServer.Address(),
		Format:       "slin48",
		ChannelID:    "managed-" + string(metadata.Direction) + "-" + h.ID(),
	})
	if err != nil {
		log.Printf("Error making %s AudioSocket: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		return
	}
	log.Printf("%s AudioSocket created: %v", metadata.Direction, mediaChannel.Key())
	bridge, err := cl.Bridge().Create(ari.NewKey(ari.BridgeKey, uuid.New().String()), "mixing", "managed-"+string(metadata.Direction)+"-"+h.ID())
	if err != nil {
		log.Printf("Error creating %s Bridge: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	err = bridge.AddChannel(snoopChannel.ID())
	if err != nil {
		log.Printf("Error adding %s channel to bridge: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	err = bridge.AddChannel(mediaChannel.ID())
	if err != nil {
		log.Printf("Error adding %s media channel to bridge: %v", metadata.Direction, err)
		err = cl.Channel().Hangup(h.Key(), "")
		err = cl.Channel().Hangup(mediaChannel.Key(), "")
		return
	}
	inMonitor := bridgemon.New(bridge)
	inEvents := inMonitor.Watch()
	*counter++
	go func() {
		log.Printf("%s Bridge Monitor started", metadata.Direction)
		for {
			m, ok := <-inEvents

			if !ok {
				log.Printf("%s Bridge Monitor closed", metadata.Direction)
				return
			}
			log.Printf("%s Got event: %v", metadata.Direction, m)

			if len(m.Channels()) <= 1 {
				err = cl.Channel().Hangup(mediaChannel.Key(), "")
				err = cl.Bridge().Delete(bridge.Key())
				rtpServer.Close()
				*counter--
				if *counter == 0 {
					audioFile, err := processAudio(metadata.Uuid)
					if err == nil {
						err = webhook.UploadFile(audioFile)
						if err != nil {
							log.Printf("Error sending audio: %v", err)
						}

					} else {
						log.Printf("Error processing audio: %v", err)
					}
				}
			}
		}
	}()
}
func processAudio(callUuid string) (string, error) {
	outputFile := "/tmp/" + callUuid + ".ogg"
	cmd := exec.Command("sox", "-M", "-r", "48000", "-e", "signed-integer", "-c", "1", "-B", "-b", "16", "/tmp/"+callUuid+"-in.raw", "-r", "48000", "-e", "signed-integer", "-c", "1", "-B", "-b", "16", "/tmp/"+callUuid+"-out.raw", outputFile)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error Running sox: %v", err)
		return "", err
	} else {
		log.Printf("Wrote file: %s", callUuid)
		os.Remove("/tmp/" + callUuid + "-in.raw")
		os.Remove("/tmp/" + callUuid + "-out.raw")

	}
	return outputFile, nil
}
