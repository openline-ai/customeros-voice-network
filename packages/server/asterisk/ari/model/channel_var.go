package model

import (
	"encoding/json"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/ghettovoice/gosip/sip/parser"
	"log"
)

type ChannelVar struct {
	Uuid                  string
	Dest                  string
	KamailioIP            string
	EndpointName          string
	OrigEndpointName      string
	OriginCarrier         *string
	DestCarrier           *string
	CallEventWebhook      *string
	RecordingEventWebhook *string
	WebookApiKey          *string
	VoiceMailPrompt       *string
	From                  *CallEventParty
	To                    *CallEventParty
}

func GetChannelVars(h *ari.ChannelHandle, hasBLeg bool) (*ChannelVar, error) {
	callUuid, err := h.GetVariable("UUID")
	if err != nil {
		log.Printf("Missing channel var UUID: %v", err)
		return nil, err
	}
	dest, err := h.GetVariable("DEST")
	if err != nil {
		log.Printf("Missing channel var DEST: %v", err)
		return nil, err
	}
	kamailioIP, err := h.GetVariable("KAMAILIO_IP")
	if err != nil {
		log.Printf("Missing channel var KAMAILIO_IP: %v", err)
		return nil, err
	}
	endpointName, err := h.GetVariable("ENDPOINT_NAME")
	if err != nil {
		log.Printf("Missing channel var ENDPOINT_NAME: %v", err)
		return nil, err
	}
	originCarrier, err := h.GetVariable("ORIGIN_CARRIER")
	var originCarrierPtr *string = nil
	if err == nil {
		originCarrierPtr = &originCarrier
	}

	destCarrier, err := h.GetVariable("DEST_CARRIER")
	var destCarrierPtr *string = nil
	if err == nil {
		destCarrierPtr = &destCarrier
	}
	origEndpointName, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Endpoint-Type)")
	if err != nil {
		log.Printf("Missing channel var PJSIP_HEADER(read,X-Openline-Endpoint-Type): %v", err)
		return nil, err
	}
	var destEndpointName string
	if hasBLeg {
		destEndpointName, err = h.GetVariable("PJSIP_HEADER(read,X-Openline-Dest-Endpoint-Type)")
		if err != nil {
			log.Printf("Missing channel var PJSIP_HEADER(read,X-Openline-Dest-Endpoint-Type): %v", err)
			return nil, err
		}
	}
	fromId := &CallEventParty{}
	fromUser, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-User)")
	if err != nil {
		fromUser = ""
	}
	from, err := h.GetVariable("PJSIP_HEADER(read,From)")
	if err != nil {
		log.Printf("Missing channel var PJSIP_HEADER(read,From): %v", err)
		return nil, err
	}
	_, fromUri, _, err := parser.ParseAddressValue(from)
	if err != nil {
		log.Printf("Error parsing From header: %v", err)
		return nil, err
	}

	if origEndpointName == "webrtc" {
		fromIdStr := fromUri.User().String() + "@" + fromUri.Host()
		fromId.Mailto = &fromIdStr
		fromId.Type = CALL_EVENT_TYPE_WEBTRC
	} else if fromUser != "" {
		base := fromUser[4:]
		fromId.Mailto = &base
		fromIdStr := fromUri.User().String() + "@" + fromUri.Host()
		fromId.Sip = &fromIdStr
		fromId.Type = CALL_EVENT_TYPE_SIP
	} else {
		fromIdStr := fromUri.User().String()
		fromId.Tel = &fromIdStr
		fromId.Type = CALL_EVENT_TYPE_PSTN
	}

	toId := &CallEventParty{}
	toUser, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Dest-User)")
	if err != nil {
		toUser = ""
	}

	to, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Dest)")
	if err != nil {
		log.Printf("Missing channel var PJSIP_HEADER(read,X-Openline-Dest): %v", err)
		return nil, err
	}

	toUri, err := parser.ParseUri(to)
	if err != nil {
		log.Printf("Error parsing To header: %v", err)
		return nil, err
	}

	if hasBLeg {
		if destEndpointName == "webrtc" {
			toStr := toUri.User().String() + "@" + toUri.Host()
			toId.Mailto = &toStr
			toId.Type = CALL_EVENT_TYPE_WEBTRC
		} else if toUser != "" && destCarrierPtr == nil {
			base := toUser[4:]
			toId.Mailto = &base
			toStr := toUri.User().String() + "@" + toUri.Host()
			toId.Sip = &toStr
			toId.Type = CALL_EVENT_TYPE_SIP
		} else {
			toStr := toUri.User().String()
			toId.Tel = &toStr
			toId.Type = CALL_EVENT_TYPE_PSTN
			if toUser != "" {
				// call forwarding, also not the user originally called
				base := toUser[4:]
				toId.Mailto = &base
			}
		}
	} else {
		toId.Type = CALL_EVENT_TYPE_VOICEMAIL
		if toUser != "" {
			// voicemail, indicate the user the message was left for
			base := toUser[4:]
			toId.Mailto = &base
		}
	}

	var recordingEventWebhook *string = nil
	var callEventWebhook *string = nil
	var webhookApiKey *string = nil

	profileInfo, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Profile-Info)")
	if err == nil {
		profile := make(map[string]string)
		err := json.Unmarshal([]byte(profileInfo), &profile)
		if err != nil {
			log.Printf("Error parsing profile info: %v", err)
			return nil, err
		}
		if callWebHook, ok := profile["call_webhook"]; ok && callWebHook != "" {
			callEventWebhook = &callWebHook
		}
		if recordingWebHook, ok := profile["recording_webhook"]; ok && recordingWebHook != "" {
			recordingEventWebhook = &recordingWebHook
		}
		if apiKey, ok := profile["api_key"]; ok && apiKey != "" {
			webhookApiKey = &apiKey
		}
	}

	var voicemailPromptObject *string = nil
	voicemailInfo, err := h.GetVariable("PJSIP_HEADER(read,X-Openline-Voicemail-Info)")
	if err == nil {
		voicemail := make(map[string]interface{})
		err := json.Unmarshal([]byte(voicemailInfo), &voicemail)
		if err != nil {
			log.Printf("Error parsing profile info: %v", err)
			return nil, err
		}
		if promptObject, ok := voicemail["prompt_object_id"]; ok && promptObject != "" {
			promptStr, ok := promptObject.(string)
			if !ok {
				log.Printf("Error parsing voicemail prompt object id: %v", err)
				return nil, err
			}
			voicemailPromptObject = &promptStr
		}
	}

	return &ChannelVar{Uuid: callUuid,
		Dest:                  dest,
		KamailioIP:            kamailioIP,
		EndpointName:          endpointName,
		OrigEndpointName:      origEndpointName,
		OriginCarrier:         originCarrierPtr,
		DestCarrier:           destCarrierPtr,
		From:                  fromId,
		To:                    toId,
		CallEventWebhook:      callEventWebhook,
		RecordingEventWebhook: recordingEventWebhook,
		WebookApiKey:          webhookApiKey,
		VoiceMailPrompt:       voicemailPromptObject,
	}, nil
}
