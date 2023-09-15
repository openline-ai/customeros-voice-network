package routes

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	awsSes "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// GetVoiceMail @Description gets a voicemail record
type GetVoiceMail struct {
	// ID of the voicemail
	ID int `json:"ID"`
	// Description of the VoiceMail entry
	Description string `json:"description" example:"VoiceMail Prompt of Agent Smith"`
	// ObjectID of the voicemail.
	ObjectID string `json:"object_uid" example:"default"`
	// Time in Seconds before forwarding failing over to voicemail
	Timeout int `json:"timeout" example:"15"`
}

type voiceMailRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /voicemail/{id} [get]
// @security ApiKeyAuth
// @Description gets a voicemail record
// @Tags		voicemail
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the voicemail entry"
// @Success     200      {object}	GetVoiceMail
// @Failure     400		 {object}	HTTPError
func (vr *voiceMailRoute) getVoiceMail(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	voiceMail, err := vr.client.OpenlineVoiceMail.Get(c, pid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	var response GetVoiceMail
	response.ID = voiceMail.ID
	response.ObjectID = voiceMail.ObjectID
	response.Description = voiceMail.Description
	response.Timeout = voiceMail.Timeout

	c.JSON(http.StatusOK, response)
}

// @Router      /voicemail [get]
// @security ApiKeyAuth
// @Description gets a list of voicemail records
// @Tags		voicemail
// @Accept      json
// @Produce     json
// @Success     200      {array}	GetVoiceMail
// @Failure     400		 {object}	HTTPError
func (vr *voiceMailRoute) getVoiceMailList(c *gin.Context) {

	query := vr.client.OpenlineVoiceMail.Query()

	voiceMailList, err := query.All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var resultList = make([]*GetVoiceMail, len(voiceMailList))

	for i := 0; i < len(voiceMailList); i++ {
		resultList[i] = &GetVoiceMail{
			ID:          voiceMailList[i].ID,
			ObjectID:    voiceMailList[i].ObjectID,
			Description: voiceMailList[i].Description,
			Timeout:     voiceMailList[i].Timeout,
		}
	}
	c.JSON(http.StatusOK, resultList)
}

// @Router      /voicemail [post]
// @security ApiKeyAuth
// @Description creates a new voicemail
// @Tags		voicemail
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       description formData string false "Voicemail Prompt Description"
// @Param       timeout formData int false "Voicemail timeout (in seconds)" default(15)
// @Param       file formData file true "Voicemail audio file"
// @Success     200      {object}	GetVoiceMail
// @Failure     400		 {object}	HTTPError
func (vr *voiceMailRoute) addVoiceMail(c *gin.Context) {
	formFile, err := c.FormFile("file")
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	description, found := c.GetPostForm("description")
	if !found {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	timeoutStr, found := c.GetPostForm("description")
	if !found {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	fileHandler, err := formFile.Open()
	if err != nil {
		NewError(c, http.StatusInternalServerError, err)
		return
	}
	defer fileHandler.Close()

	session, err := awsSes.NewSession(&aws.Config{Region: aws.String(vr.conf.AWS.Region)})
	if err != nil {
		NewError(c, http.StatusInternalServerError, err)
		return
	}

	tempFile, err := createTempFileWithExtension(formFile.Filename)
	if err != nil {
		NewError(c, http.StatusInternalServerError, fmt.Errorf("addVoiceMail: failed to create temp file: %v", err))
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, fileHandler)
	if err != nil {
		NewError(c, http.StatusInternalServerError, fmt.Errorf("addVoiceMail: failed to copy file: %v", err))
		tempFile.Close()
		return
	}
	tempFile.Close()

	transcodedFile, err := transcodeAudioFile(tempFile.Name())
	defer os.Remove(transcodedFile)
	if err != nil {
		NewError(c, http.StatusInternalServerError, fmt.Errorf("addVoiceMail: failed to transcode file: %v", err))
		return
	}
	tf, err := os.Open(transcodedFile)
	if err != nil {
		NewError(c, http.StatusInternalServerError, fmt.Errorf("addVoiceMail: failed to open transcoded file: %v", err))
		return
	}
	defer tf.Close()

	objectID, err := uuid.NewUUID()
	if err != nil {
		NewError(c, http.StatusInternalServerError, err)
		return
	}

	transInfo, err := os.Stat(transcodedFile)
	if err != nil {
		NewError(c, http.StatusInternalServerError, fmt.Errorf("addVoiceMail: failed to get transcoded file info: %v", err))
		return
	}

	input := &s3.PutObjectInput{
		Bucket:               aws.String(vr.conf.AWS.Bucket),
		Key:                  aws.String(objectID.String()),
		ACL:                  aws.String("private"),
		Body:                 tf,
		ContentLength:        aws.Int64(transInfo.Size()),
		ContentType:          aws.String("audio/wav"),
		ContentDisposition:   aws.String(fmt.Sprintf("attachment; filename=%s", filepath.Base(transcodedFile))),
		ServerSideEncryption: aws.String("AES256"),
	}
	_, objErr := s3.New(session).PutObject(input)
	if objErr != nil {
		NewError(c, http.StatusInternalServerError, objErr)
		return
	}

	voiceMail, err := vr.client.OpenlineVoiceMail.Create().
		SetObjectID(objectID.String()).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		SetEnabled(true).
		SetTimeout(timeout).
		SetDescription(description).
		Save(c)

	if err != nil {
		vr.deleteVoicemailFromS3(session, objectID.String())
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetVoiceMail
	response.ID = voiceMail.ID
	response.ObjectID = voiceMail.ObjectID
	response.Description = voiceMail.Description
	c.JSON(http.StatusOK, response)
}

func transcodeAudioFile(inputFile string) (string, error) {
	tempFile, err := os.CreateTemp("", "*.wav")
	if err != nil {
		return "", fmt.Errorf("transcodeAudioFile: Unable to create Tmp File %v", err)
	}
	tempFile.Close()
	cmd := exec.Command("sox", inputFile, "-r", "8000", "-c", "1", "-e", "signed-integer", "-b", "16", tempFile.Name())
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("transcodeAudioFile: Unable to transcode file %v", err)
	}
	return tempFile.Name(), nil
}

func createTempFileWithExtension(filename string) (*os.File, error) {
	extension := filepath.Ext(filename)
	tempFile, err := os.CreateTemp("", "*"+extension)
	if err != nil {
		return nil, err
	}
	return tempFile, nil
}

func (vr *voiceMailRoute) deleteVoicemailFromS3(session *awsSes.Session, objectID string) error {
	if session == nil {
		var err error
		session, err = awsSes.NewSession(&aws.Config{Region: aws.String(vr.conf.AWS.Region)})
		if err != nil {
			return err
		}
	}

	// delete file from s3
	_, s3err := s3.New(session).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(vr.conf.AWS.Bucket),
		Key:    aws.String(objectID),
	})
	if s3err != nil {
		return fmt.Errorf("deleteVoiceMailFromS3: %v", s3err)
	}
	return nil
}

// @Router      /voicemail/{id} [delete]
// @security ApiKeyAuth
// @Description deletes a voiceMail record
// @Tags		voicemail
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the voiceMail entry"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (vr *voiceMailRoute) deleteVoiceMail(c *gin.Context) {
	id := c.Param("id")
	pid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	voiceMail, err := vr.client.OpenlineVoiceMail.Get(c, pid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	err = vr.deleteVoicemailFromS3(nil, voiceMail.ObjectID)
	if err != nil {
		NewError(c, http.StatusBadRequest, fmt.Errorf("Unable to delete voicemail from S3: %v", err))
		return
	}

	err = vr.client.OpenlineVoiceMail.DeleteOneID(pid).Exec(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

func addVoiceMailRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {
	pr := new(voiceMailRoute)
	pr.conf = conf
	pr.client = client

	rg.GET("voicemail/:id", pr.getVoiceMail)
	rg.GET("voicemail", pr.getVoiceMailList)
	rg.POST("voicemail", pr.addVoiceMail)
	rg.DELETE("voicemail/:id", pr.deleteVoiceMail)
}
