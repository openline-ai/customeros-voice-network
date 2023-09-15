package routes

import (
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen/openlineprofile"
	"net/http"
	"strconv"
)

// @Description defines the set of webhooks for a profile
type GetProfile struct {
	// ID of the profile
	ID int `json:"ID"`
	// Name of the profile
	ProfileName string `json:"profile_name" example:"default"`
	// Webhook to call when a call state changes, if required
	CallWebhook string `json:"call_webhook" example:"https://myserver.com/call_webhook"`
	// Webhook to call when send recordings, if required
	RecordingWebhook string `json:"recording_webhook" example:"https://myserver.com/recording_webhook"`
	// API key to send to webhooks, if required
	ApiKey string `json:"api_key" example:"my_api_key"`
}

// @Description Identical to GetProfile except ID is omittied
type AddProfile struct {
	// Name of the profile
	ProfileName string `json:"profile_name" example:"default"`
	// Webhook to call when a call state changes, if required
	CallWebhook string `json:"call_webhook" example:"https://myserver.com/call_webhook"`
	// Webhook to call when send recordings, if required
	RecordingWebhook string `json:"recording_webhook" example:"https://myserver.com/recording_webhook"`
	// API key to send to webhooks, if required
	ApiKey string `json:"api_key" example:"my_api_key"`
}

type profileRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /profile/{id} [get]
// @security ApiKeyAuth
// @Description gets a profile record
// @Tags		profile
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the number mapping entry"
// @Success     200      {object}	GetProfile
// @Failure     400		 {object}	HTTPError
func (pr *profileRoute) getProfile(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	profile, err := pr.client.OpenlineProfile.Get(c, pid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	var response GetProfile
	response.ID = profile.ID
	response.ProfileName = profile.ProfileName
	response.CallWebhook = profile.CallWebhook
	response.RecordingWebhook = profile.RecordingWebhook
	response.ApiKey = profile.APIKey
	c.JSON(http.StatusOK, response)
}

// @Router      /profile [get]
// @security ApiKeyAuth
// @Description gets a list of profile records
// @Tags		profile
// @Accept      json
// @Produce     json
// @Param       profile_name  query     string false "name of profile to filter for" example(openline_profile)
// @Success     200      {array}	GetProfile
// @Failure     400		 {object}	HTTPError
func (pr *profileRoute) getProfileList(c *gin.Context) {
	profileName := c.Query("profile_name")
	query := pr.client.OpenlineProfile.Query()

	if profileName != "" {
		query = query.Where(openlineprofile.ProfileNameEQ(profileName))
	}

	profileList, err := query.All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var resultList = make([]*GetProfile, len(profileList))

	for i := 0; i < len(profileList); i++ {
		resultList[i] = &GetProfile{
			ID:               profileList[i].ID,
			ProfileName:      profileList[i].ProfileName,
			CallWebhook:      profileList[i].CallWebhook,
			RecordingWebhook: profileList[i].RecordingWebhook,
			ApiKey:           profileList[i].APIKey,
		}
	}
	c.JSON(http.StatusOK, resultList)
}

// @Router      /profile [post]
// @security ApiKeyAuth
// @Description creates a new profile
// @Tags		profile
// @Accept      json
// @Produce     json
// @Param       message  body  AddProfile false "profile to insert into the database"
// @Success     200      {object}	GetProfile
// @Failure     400		 {object}	HTTPError
func (pr *profileRoute) addProfile(c *gin.Context) {
	var newProfile AddProfile
	if err := c.ShouldBindJSON(&newProfile); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	profile, err := pr.client.OpenlineProfile.Create().
		SetProfileName(newProfile.ProfileName).
		SetCallWebhook(newProfile.CallWebhook).
		SetRecordingWebhook(newProfile.RecordingWebhook).
		SetAPIKey(newProfile.ApiKey).
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetProfile
	response.ID = profile.ID
	response.ProfileName = profile.ProfileName
	response.CallWebhook = profile.CallWebhook
	response.RecordingWebhook = profile.RecordingWebhook
	response.ApiKey = profile.APIKey
	c.JSON(http.StatusOK, response)
}

// @Router      /profile/{id} [put]
// @security ApiKeyAuth
// @Description updates the specified profile
// @Tags		profile
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the profile entry"
// @Param       message  body  AddProfile false "revised profile to update the database with"
// @Success     200      {object}	GetProfile
// @Failure     400		 {object}	HTTPError
func (pr *profileRoute) updateProfile(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	var newProfile AddProfile
	if err := c.ShouldBindJSON(&newProfile); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	profile, err := pr.client.OpenlineProfile.UpdateOneID(pid).
		SetProfileName(newProfile.ProfileName).
		SetCallWebhook(newProfile.CallWebhook).
		SetRecordingWebhook(newProfile.RecordingWebhook).
		SetAPIKey(newProfile.ApiKey).
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetProfile
	response.ID = profile.ID
	response.ProfileName = profile.ProfileName
	response.CallWebhook = profile.CallWebhook
	response.RecordingWebhook = profile.RecordingWebhook
	response.ApiKey = profile.APIKey
	c.JSON(http.StatusOK, response)
}

// @Router      /profile/{id} [delete]
// @security ApiKeyAuth
// @Description deletes a profile record
// @Tags		profile
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the profile entry"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (pr *profileRoute) deleteProfile(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	err = pr.client.OpenlineProfile.DeleteOneID(pid).Exec(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

func addProfileRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {
	pr := new(profileRoute)
	pr.conf = conf
	pr.client = client

	rg.GET("profile/:id", pr.getProfile)
	rg.GET("profile", pr.getProfileList)
	rg.POST("profile", pr.addProfile)
	rg.PUT("profile/:id", pr.updateProfile)
	rg.DELETE("profile/:id", pr.deleteProfile)

}
