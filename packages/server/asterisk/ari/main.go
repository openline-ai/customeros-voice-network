package main

import (
	"agi/server"
	"agi/services"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/client/native"
	"gopkg.in/ini.v1"
	"log"
	"strings"
)

func main() {
	ariCfg, err := ini.Load("/etc/asterisk/ari.conf")
	if err != nil {
		log.Fatal("Unable to read asterisk config file")
	}

	awsCfg, err := ini.Load("/etc/asterisk/aws.conf")
	if err != nil {
		log.Fatal("Unable to read AWS config file")
	}
	go startRecordAri(ariCfg)
	startVoicemailAri(ariCfg, awsCfg)

}

func startVoicemailAri(ariCfg *ini.File, awsCfg *ini.File) {

	voicemailClient, err := native.Connect(&native.Options{
		Application:  "voicemail",
		Username:     "asterisk",
		Password:     ariCfg.Section("asterisk").Key("password").String(),
		URL:          "http://localhost:8088/ari",
		WebsocketURL: "ws://localhost:8088/ari/events",
	})

	if err != nil {
		log.Fatalf("Unable to create ari server %v", err)
	}
	log.Printf("Voicemail Asterisk ARI client created")
	log.Printf("Listening for new calls")
	voicemailSub := voicemailClient.Bus().Subscribe(nil, "StasisStart")

	voicemailAriHandler := server.NewVoicemailAriHandler(services.ServiceFactoryImpl{})
	for {
		select {
		case e := <-voicemailSub.Events():
			v := e.(*ari.StasisStart)
			log.Printf("Voicemail: Got stasis start channel: %s", v.Channel.ID)
			if !strings.HasPrefix(v.Channel.ID, "managed") {
				go voicemailAriHandler.App(voicemailClient, voicemailClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)), awsCfg)
			}
		}
	}
}

func startRecordAri(cfg *ini.File) {
	recordingClient, err := native.Connect(&native.Options{
		Application:  "recording",
		Username:     "asterisk",
		Password:     cfg.Section("asterisk").Key("password").String(),
		URL:          "http://localhost:8088/ari",
		WebsocketURL: "ws://localhost:8088/ari/events",
	})

	if err != nil {
		log.Fatalf("Unable to create ari server %v", err)
	}
	log.Printf("Recording Asterisk ARI client created")
	log.Printf("Listening for new calls")
	recordingSub := recordingClient.Bus().Subscribe(nil, "StasisStart")

	recordAriHandler := server.NewRecordAriHandler(services.ServiceFactoryImpl{})
	for {
		select {
		case e := <-recordingSub.Events():
			v := e.(*ari.StasisStart)
			log.Printf("Recording: Got stasis start channel: %s", v.Channel.ID)
			if !strings.HasPrefix(v.Channel.ID, "managed") {
				go recordAriHandler.App(recordingClient, recordingClient.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)))
			}
		}
	}

}
