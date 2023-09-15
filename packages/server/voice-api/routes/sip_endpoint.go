package routes

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen/kamailiosubscriber"
	"net/http"
	"strconv"
)

// @Description: Information required for sip endpoints to register (including esims)
type GetSipEndpoint struct {
	// the id for this specific sip endpoint record
	ID int `json:"ID"`
	// username the endpoint sends when registering
	Username string `json:"username"  example:"my-cool-esim"`
	// domain the endpoint puts in the From header when registering
	Domain string `json:"domain"  example:"openline.ai"`
}

// @Description: Information required for sip endpoints to register (including esims)
type AddSipEndpoint struct {
	// username the endpoint sends when registering
	Username string `json:"username"  example:"my-cool-esim"`
	// domain the endpoint puts in the From header when registering
	Domain string `json:"domain"  example:"openline.ai"`
	// password used by the endpoint to authenticate with us
	Password string `json:"password"  example:"my-secret-password"`
}

type sipEndpointRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /sip_endpoint/{id} [get]
// @security ApiKeyAuth
// @Description gets a sip endpoint record
// @Tags		sip_endpoint
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the sip endpoint record"
// @Success     200      {object}	GetSipEndpoint
// @Failure     400		 {object}	HTTPError
func (ser *sipEndpointRoute) getSipEndpoint(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	subscriber, err := ser.client.KamailioSubscriber.Get(c, cid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetSipEndpoint
	response.ID = subscriber.ID
	response.Domain = subscriber.Domain
	response.Username = subscriber.Username

	c.JSON(http.StatusOK, response)
}

// @Router      /sip_endpoint/{id} [delete]
// @security ApiKeyAuth
// @Description deletes a sip endpoint record
// @Tags		sip_endpoint
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the sip endpoint to delete"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (ser *sipEndpointRoute) deleteSipEndpoint(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	err = ser.client.KamailioSubscriber.DeleteOneID(cid).Exec(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

// @Router      /sip_endpoint [get]
// @security ApiKeyAuth
// @Description gets a list sip endpoint records
// @Tags		sip_endpoint
// @Accept      json
// @Produce     json
// @Param       username  query     string false "username to filter sip endpoint set against" example(my-cool-esim)
// @Param       domain  query     string false "domain to filter sip endpoint set against" example(openline.ai)
// @Success     200      {array}	GetSipEndpoint
// @Failure     400		 {object}	HTTPError
func (ser *sipEndpointRoute) getSipEndpointList(c *gin.Context) {
	username := c.Query("username")
	domain := c.Query("domain")

	query := ser.client.KamailioSubscriber.Query()

	if username != "" {
		query = query.Where(kamailiosubscriber.UsernameEQ(username))
	}

	if domain != "" {
		query = query.Where(kamailiosubscriber.DomainEQ(domain))
	}

	sipEndpointList, err := query.All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	resultList := make([]*GetSipEndpoint, len(sipEndpointList))

	for i := 0; i < len(sipEndpointList); i++ {
		resultList[i] = &GetSipEndpoint{
			ID:       sipEndpointList[i].ID,
			Domain:   sipEndpointList[i].Domain,
			Username: sipEndpointList[i].Username,
		}
	}

	c.JSON(http.StatusOK, &resultList)
}

// @Router      /sip_endpoint [post]
// @security ApiKeyAuth
// @Description creates a new sip endpoint record
// @Tags		sip_endpoint
// @Accept      json
// @Produce     json
// @Param       message  body  AddSipEndpoint false "sip endpoint record to insert into the database"
// @Success     200      {object}	GetSipEndpoint
// @Failure     400		 {object}	HTTPError
func (ser *sipEndpointRoute) addSipEndpoint(c *gin.Context) {
	var newSipEndpoint AddSipEndpoint
	if err := c.ShouldBindJSON(&newSipEndpoint); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	data := md5.Sum([]byte(newSipEndpoint.Username + ":" + newSipEndpoint.Domain + ":" + newSipEndpoint.Password))
	ha1 := fmt.Sprintf("%x", data)

	data = md5.Sum([]byte(newSipEndpoint.Username + "@" + newSipEndpoint.Domain + ":" + newSipEndpoint.Domain + ":" + newSipEndpoint.Password))
	ha1b := fmt.Sprintf("%x", data)

	sipEndpoint, err := ser.client.KamailioSubscriber.Create().
		SetUsername(newSipEndpoint.Username).
		SetDomain(newSipEndpoint.Domain).
		SetHa1(ha1).
		SetHa1b(ha1b).
		Save(c)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetSipEndpoint
	response.ID = sipEndpoint.ID
	response.Username = sipEndpoint.Username
	response.Domain = sipEndpoint.Domain

	c.JSON(http.StatusOK, &response)
}

func addSipEndpointRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {

	ser := new(sipEndpointRoute)
	ser.conf = conf
	ser.client = client

	rg.GET("sip_endpoint/:id", ser.getSipEndpoint)
	rg.DELETE("sip_endpoint/:id", ser.deleteSipEndpoint)
	rg.GET("sip_endpoint", ser.getSipEndpointList)
	rg.POST("sip_endpoint", ser.addSipEndpoint)

}
