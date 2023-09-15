package routes

import (
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"net/http"
	"strconv"
)

// @Description defines the number to forward calls to
type GetForwarding struct {
	// ID of the forwarding
	ID int `json:"ID"`
	// Description of the forwarding
	Description string `json:"description" example:"Agent Smith's Mobile"`
	// True if traffic is to be forwarded, false otherwise
	Enabled bool `json:"enabled" example:"true"`
	// Number to forward calls to
	E164 string `json:"e164" example:"+15551234567"`
}

// @Description Identical to GetForwarding except ID is omittied
type AddForwarding struct {
	// Name of the forwarding
	Description string `json:"description" example:"Agent Smith's Mobile"`
	// True if traffic is to be forwarded, false otherwise
	Enabled bool `json:"enabled" example:"true"`
	// Number to forward calls to
	E164 string `json:"e164" example:"+15551234567"`
}

type forwardingRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /forwarding/{id} [get]
// @security ApiKeyAuth
// @Description gets a forwarding record
// @Tags		forwarding
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the forwarding entry"
// @Success     200      {object}	GetForwarding
// @Failure     400		 {object}	HTTPError
func (fr *forwardingRoute) getForwarding(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	forwarding, err := fr.client.OpenlineForwarding.Get(c, pid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	var response GetForwarding
	response.ID = forwarding.ID
	response.Description = forwarding.Description
	response.Enabled = forwarding.Enabled
	response.E164 = forwarding.E164

	c.JSON(http.StatusOK, response)
}

// @Router      /forwarding [get]
// @security ApiKeyAuth
// @Description gets a list of forwardings
// @Tags		forwarding
// @Accept      json
// @Produce     json
// @Success     200      {array}	GetForwarding
// @Failure     400		 {object}	HTTPError
func (fr *forwardingRoute) getForwardingList(c *gin.Context) {
	query := fr.client.OpenlineForwarding.Query()

	forwardingList, err := query.All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var resultList = make([]*GetForwarding, len(forwardingList))

	for i := 0; i < len(forwardingList); i++ {
		resultList[i] = &GetForwarding{
			ID:          forwardingList[i].ID,
			Description: forwardingList[i].Description,
			Enabled:     forwardingList[i].Enabled,
			E164:        forwardingList[i].E164,
		}
	}
	c.JSON(http.StatusOK, resultList)
}

// @Router      /forwarding [post]
// @security ApiKeyAuth
// @Description creates a new forwarding
// @Tags		forwarding
// @Accept      json
// @Produce     json
// @Param       message  body  AddForwarding false "forwarding to insert into the database"
// @Success     200      {object}	GetForwarding
// @Failure     400		 {object}	HTTPError
func (fr *forwardingRoute) addForwarding(c *gin.Context) {
	var newForwarding AddForwarding
	if err := c.ShouldBindJSON(&newForwarding); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	forwarding, err := fr.client.OpenlineForwarding.Create().
		SetDescription(newForwarding.Description).
		SetEnabled(newForwarding.Enabled).
		SetE164(newForwarding.E164).
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetForwarding
	response.ID = forwarding.ID
	response.Description = forwarding.Description
	response.Enabled = forwarding.Enabled
	response.E164 = forwarding.E164
	c.JSON(http.StatusOK, response)
}

// @Router      /forwarding/{id} [put]
// @security ApiKeyAuth
// @Description updates the specified forwarding
// @Tags		forwarding
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the forwarding entry"
// @Param       message  body  AddForwarding false "revised forwarding to update the database with"
// @Success     200      {object}	GetForwarding
// @Failure     400		 {object}	HTTPError
func (fr *forwardingRoute) updateForwarding(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	var newForwarding AddForwarding
	if err := c.ShouldBindJSON(&newForwarding); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	forwarding, err := fr.client.OpenlineForwarding.UpdateOneID(pid).
		SetDescription(newForwarding.Description).
		SetEnabled(newForwarding.Enabled).
		SetE164(newForwarding.E164).
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetForwarding
	response.ID = forwarding.ID
	response.Description = forwarding.Description
	response.Enabled = forwarding.Enabled
	response.E164 = forwarding.E164
	c.JSON(http.StatusOK, response)
}

// @Router      /forwarding/{id} [delete]
// @security ApiKeyAuth
// @Description deletes a forwarding record
// @Tags		forwarding
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the forwarding entry"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (fr *forwardingRoute) deleteForwarding(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	err = fr.client.OpenlineForwarding.DeleteOneID(pid).Exec(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

func addForwardingRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {
	fr := new(forwardingRoute)
	fr.conf = conf
	fr.client = client

	rg.GET("forwarding/:id", fr.getForwarding)
	rg.GET("forwarding", fr.getForwardingList)
	rg.POST("forwarding", fr.addForwarding)
	rg.PUT("forwarding/:id", fr.updateForwarding)
	rg.DELETE("forwarding/:id", fr.deleteForwarding)

}
