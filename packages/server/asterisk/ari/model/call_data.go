package model

type CallEventPartyType string

const (
	CALL_EVENT_TYPE_PSTN      = CallEventPartyType("pstn")
	CALL_EVENT_TYPE_SIP       = CallEventPartyType("sip")
	CALL_EVENT_TYPE_WEBTRC    = CallEventPartyType("webrtc")
	CALL_EVENT_TYPE_VOICEMAIL = CallEventPartyType("voicemail")
)

type CallEventParty struct {
	Tel    *string            `json:"tel,omitempty"`
	Mailto *string            `json:"mailto,omitempty"`
	Sip    *string            `json:"sip,omitempty"`
	Type   CallEventPartyType `json:"type"`
}

type CallMetadata struct {
	From      *CallEventParty
	To        *CallEventParty
	Tenant    string
	Uuid      string
	Direction CallDirection
}

type CallDirection string

const (
	IN  CallDirection = "in"
	OUT CallDirection = "out"
)

func MakeMetaData(direction CallDirection, vars *ChannelVar) *CallMetadata {
	return &CallMetadata{
		Direction: direction,
		Uuid:      vars.Uuid,
		From:      vars.From,
		To:        vars.To,
	}
}
