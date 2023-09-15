package server

import (
	"agi/model"
	"agi/services"
	"context"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/ext/play"
	"github.com/CyCoreSystems/ari/v6/ext/record"
	"gopkg.in/ini.v1"
	"log"
	"os"
)

type VoicemailAriHandler struct {
	serviceFactory services.ServiceFactory
}

func NewVoicemailAriHandler(serviceFactory services.ServiceFactory) *VoicemailAriHandler {
	return &VoicemailAriHandler{
		serviceFactory: serviceFactory,
	}
}

func (ah *VoicemailAriHandler) App(cl ari.Client, h *ari.ChannelHandle, awsCfg *ini.File) {
	log.Printf("Running Voicemail App channel: %s", h.Key().ID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aHangup := h.Subscribe(ari.Events.StasisEnd)

	channelVars, err := model.GetChannelVars(h, false)
	if err != nil {
		log.Printf("Error getting channel vars: %v", err)
		err = cl.Channel().Hangup(h.Key(), "")
		h.Busy()
		return
	}

	webhookClient := ah.serviceFactory.NewWebhookClient(channelVars)

	webhookClient.StartCallEvent()

	if channelVars.VoiceMailPrompt == nil {
		log.Printf("No voicemail prompt set for %s", channelVars.Dest)
		h.Busy()
		return
	}

	s3client := ah.serviceFactory.NewS3Client(awsCfg)

	url, err := s3client.GetUrl(*channelVars.VoiceMailPrompt)

	if err != nil {
		log.Printf("Error getting voicemail url: %v", err)
		h.Busy()
		return
	}

	log.Printf("Voicemail url: %s", url)

	if err := h.Answer(); err != nil {
		log.Printf("failed to answer call: %v", err)
		h.Busy()
		return
	}
	// End the app when the channel goes away
	go func() {
		<-aHangup.Events()
		cancel()
	}()

	if err := play.Play(ctx, h, play.URI("sound:"+url)).Err(); err != nil {
		log.Printf("failed to play sound %v", err)
		h.Hangup()
		return
	}

	webhookClient.AnwswerCallEvent()

	res, err := record.Record(ctx, h,
		record.TerminateOn("any"),
		record.IfExists("overwrite"),
		record.Beep(),
	).Result()

	if err != nil {
		log.Printf("failed to record call: %v", err)
		return
	}

	res.Save(channelVars.Uuid)

	webhookClient.EndCallEvent(true)
	webhookClient.UploadFile("/var/spool/asterisk/recording/" + channelVars.Uuid + ".wav")
	if err := os.Remove("/var/spool/asterisk/recording/" + channelVars.Uuid + ".wav"); err != nil {
		log.Printf("failed to remove recording file: %v", err)
		return
	}
}
