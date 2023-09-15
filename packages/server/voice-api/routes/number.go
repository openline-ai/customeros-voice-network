package routes

import (
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen/openlinenumbermapping"
	"net/http"
	"strconv"
)

// @Description defines the mapping between telephone numbers and a sip address of a user in the system
type GetNumberMapping struct {
	// the id for this specific IP record
	ID int `json:"ID"`
	// Telephone number associated with the sip_uri
	E164 string `json:"e164" example:"+4412345678"`
	// Number to alias to when this sipuri makes outbound calls
	Alias string `json:"alias" example:"+4412340000"`
	// the sip uri to associate with this number
	SipUri string `json:"sip_uri" example:"sip:AgentSmith@agent.openline.ai"`
	// the uri of the esim provider to use for this number, leave blank for none
	PhoneUri string `json:"phone_uri" example:"sip:test001@tango.org"`
	// name of the carrier this number is associated with
	Carrier string `json:"carrier_name"  example:"twilio"`
	// the id of the profile this number is associated with, if any
	ProfileID *int `json:"profile_id"`
	// the voicemail box id associated with this number, if any
	VoicemailID *int `json:"voicemail_id"`
	// the forwarding id associated with this number, if any
	ForwardingID *int `json:"forwarding_id"`
}

// @Description Identical to GetNumberMapping except ID is omittied
type AddNumberMapping struct {
	// Telephone number associated with the sip_uri
	E164 string `json:"e164" example:"+4412345678"`
	// Number to alias to when this sipuri makes outbound calls
	Alias string `json:"alias" example:"+4412340000"`
	// the sip uri to associate with this number
	SipUri string `json:"sip_uri" example:"sip:AgentSmith@agent.openline.ai"`
	// the uri of the esim provider to use for this number, leave blank for none
	PhoneUri string `json:"phone_uri" example:"sip:test001@tango.org"`
	// name of the carrier this number is associated with
	Carrier string `json:"carrier_name"  example:"twilio"`
	// the id of the profile this number is associated with
	ProfileID *int `json:"profile_id"`
	// the id of the voicemail box associated with this number
	VoicemailID *int `json:"voicemail_id"`
	// the forwarding id associated with this number, if any
	ForwardingID *int `json:"forwarding_id"`
}

type numberRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /number_mapping/{id} [get]
// @security ApiKeyAuth
// @Description gets a number mapping record
// @Tags		number_mapping
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the number mapping entry"
// @Success     200      {object}	GetNumberMapping
// @Failure     400		 {object}	HTTPError
func (nr *numberRoute) getMapping(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	numberMapping, err := nr.client.OpenlineNumberMapping.Get(c, cid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetNumberMapping
	response.ID = numberMapping.ID
	response.E164 = numberMapping.E164
	response.Alias = numberMapping.Alias
	response.SipUri = numberMapping.Sipuri
	response.PhoneUri = numberMapping.Phoneuri
	response.Carrier = numberMapping.CarrierName
	response.ProfileID = numberMapping.ProfileID
	response.VoicemailID = numberMapping.VoicemailID
	response.ForwardingID = numberMapping.ForwardingID
	c.JSON(http.StatusOK, response)
}

// @Router      /number_mapping/{id} [delete]
// @security ApiKeyAuth
// @Description deletes a number mapping record
// @Tags		number_mapping
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the number mapping entry"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (nr *numberRoute) deleteMapping(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	err = nr.client.OpenlineNumberMapping.DeleteOneID(cid).Exec(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

// @Router      /number_mapping [get]
// @security ApiKeyAuth
// @Description gets a list of number mapping records
// @Tags		number_mapping
// @Accept      json
// @Produce     json
// @Param       carrier_name  query     string false "carrier to filter number mapping list against" example(twillio)
// @Param       e164  		  query     string false "e164 to find" example(+4412345678)
// @Param       sip_uri  	  query     string false "sip_uri to find" example(sip:AgentSmith@agent.openline.ai)
// @Param       profile_id    query     string false "profile_id to find" example(1)
// @Success     200      {array}	GetNumberMapping
// @Failure     400		 {object}	HTTPError
func (nr *numberRoute) getMappingList(c *gin.Context) {
	carrierName := c.Query("carrier_name")
	e164 := c.Query("e164")
	sipUri := c.Query("sip_uri")
	profileId := c.Query("profile_id")

	query := nr.client.OpenlineNumberMapping.Query()

	if carrierName != "" {
		query = query.Where(openlinenumbermapping.CarrierNameEQ(carrierName))
	}

	if e164 != "" {
		query = query.Where(openlinenumbermapping.E164(e164))
	}

	if sipUri != "" {
		query = query.Where(openlinenumbermapping.Sipuri(sipUri))
	}

	if profileId != "" {
		pid, err := strconv.Atoi(profileId)
		if err != nil {
			NewError(c, http.StatusBadRequest, err)
			return
		}
		query = query.Where(openlinenumbermapping.ProfileIDEQ(pid))
	}

	numberMappingList, err := query.All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var resultList = make([]*GetNumberMapping, len(numberMappingList))

	for i := 0; i < len(numberMappingList); i++ {
		resultList[i] = &GetNumberMapping{
			ID:           numberMappingList[i].ID,
			E164:         numberMappingList[i].E164,
			Alias:        numberMappingList[i].Alias,
			SipUri:       numberMappingList[i].Sipuri,
			PhoneUri:     numberMappingList[i].Phoneuri,
			Carrier:      numberMappingList[i].CarrierName,
			ProfileID:    numberMappingList[i].ProfileID,
			VoicemailID:  numberMappingList[i].VoicemailID,
			ForwardingID: numberMappingList[i].ForwardingID,
		}
	}
	c.JSON(http.StatusOK, resultList)
}

// @Router      /number_mapping/{id} [put]
// @security ApiKeyAuth
// @Description updates an existing number mapping
// @Tags		number_mapping
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the number mapping entry to update"
// @Param       message  body  AddNumberMapping false "number mapping to update the database with"
// @Success     200      {object}	GetNumberMapping
// @Failure     400		 {object}	HTTPError
func (nr *numberRoute) updateMapping(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var newNumber AddNumberMapping
	if err := c.ShouldBindJSON(&newNumber); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	intermediateNumberMappingUpdate := nr.client.OpenlineNumberMapping.UpdateOneID(cid).
		SetE164(newNumber.E164).
		SetAlias(newNumber.Alias).
		SetSipuri(newNumber.SipUri).
		SetPhoneuri(newNumber.PhoneUri).
		SetCarrierName(newNumber.Carrier)

	if newNumber.ProfileID != nil {
		intermediateNumberMappingUpdate = intermediateNumberMappingUpdate.SetProfileID(*newNumber.ProfileID)
	} else {
		intermediateNumberMappingUpdate = intermediateNumberMappingUpdate.ClearOpenlineProfile()
	}

	if newNumber.VoicemailID != nil {
		intermediateNumberMappingUpdate = intermediateNumberMappingUpdate.SetVoicemailID(*newNumber.VoicemailID)
	} else {
		intermediateNumberMappingUpdate = intermediateNumberMappingUpdate.ClearOpenlineVoicemail()
	}

	if newNumber.ForwardingID != nil {
		intermediateNumberMappingUpdate = intermediateNumberMappingUpdate.SetForwardingID(*newNumber.ForwardingID)
	} else {
		intermediateNumberMappingUpdate = intermediateNumberMappingUpdate.ClearOpenlineForwarding()
	}
	numberMapping, err := intermediateNumberMappingUpdate.
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetNumberMapping
	response.ID = numberMapping.ID
	response.E164 = numberMapping.E164
	response.Alias = numberMapping.Alias
	response.SipUri = numberMapping.Sipuri
	response.PhoneUri = numberMapping.Phoneuri
	response.Carrier = numberMapping.CarrierName
	response.ProfileID = numberMapping.ProfileID
	response.VoicemailID = numberMapping.VoicemailID
	response.ForwardingID = numberMapping.ForwardingID
	c.JSON(http.StatusOK, response)
}

// @Router      /number_mapping [post]
// @security ApiKeyAuth
// @Description creates a new number mapping
// @Tags		number_mapping
// @Accept      json
// @Produce     json
// @Param       message  body  AddNumberMapping false "number mapping to insert into the database"
// @Success     200      {object}	GetNumberMapping
// @Failure     400		 {object}	HTTPError
func (nr *numberRoute) addMapping(c *gin.Context) {
	var newNumber AddNumberMapping
	if err := c.ShouldBindJSON(&newNumber); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	numberMapping, err := nr.client.OpenlineNumberMapping.Create().
		SetE164(newNumber.E164).
		SetAlias(newNumber.Alias).
		SetSipuri(newNumber.SipUri).
		SetPhoneuri(newNumber.PhoneUri).
		SetCarrierName(newNumber.Carrier).
		SetNillableProfileID(newNumber.ProfileID).
		SetNillableVoicemailID(newNumber.VoicemailID).
		SetNillableForwardingID(newNumber.ForwardingID).
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetNumberMapping
	response.ID = numberMapping.ID
	response.E164 = numberMapping.E164
	response.Alias = numberMapping.Alias
	response.SipUri = numberMapping.Sipuri
	response.PhoneUri = numberMapping.Phoneuri
	response.Carrier = numberMapping.CarrierName
	response.ProfileID = numberMapping.ProfileID
	response.VoicemailID = numberMapping.VoicemailID
	response.ForwardingID = numberMapping.ForwardingID
	c.JSON(http.StatusOK, response)
}

func addNumberRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {

	nr := new(numberRoute)
	nr.conf = conf
	nr.client = client

	rg.GET("number_mapping/:id", nr.getMapping)
	rg.GET("number_mapping", nr.getMappingList)
	rg.POST("number_mapping", nr.addMapping)
	rg.PUT("number_mapping/:id", nr.updateMapping)
	rg.DELETE("number_mapping/:id", nr.deleteMapping)

}
