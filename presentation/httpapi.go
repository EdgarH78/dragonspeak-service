package presentation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Handle string `json:"handle"`
	Email  string `json:"email"`
}

func (c CreateUserRequest) toUser() models.User {
	return models.User{
		Handle: c.Handle,
		Email:  c.Email,
	}
}

type UserResponse struct {
	Handle string `json:"handle"`
	Email  string `json:"email"`
	ID     string `json:"id"`
}

func UserResponseFromUser(user *models.User) UserResponse {
	return UserResponse{
		Handle: user.Handle,
		Email:  user.Email,
		ID:     user.ID,
	}
}

type CreateCampaignRequest struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

func (c CreateCampaignRequest) toCampaign() models.Campaign {
	return models.Campaign{
		Name: c.Name,
		Link: c.Link,
	}
}

type CampaignResponse struct {
	Name string `json:"name"`
	Link string `json:"link"`
	ID   string `json:"id"`
}

func CampaignResponseFromCampaign(campaign *models.Campaign) CampaignResponse {
	return CampaignResponse{
		Name: campaign.Name,
		Link: campaign.Link,
		ID:   campaign.ID,
	}
}

type CreateSessionRequest struct {
	SessionDate time.Time `json:"sessionDate"`
	Title       string    `json:"title"`
}

func (c CreateSessionRequest) toSession() models.Session {
	return models.Session{
		SessionDate: c.SessionDate,
		Title:       c.Title,
	}
}

type SessionResponse struct {
	ID          string    `json:"id"`
	SessionDate time.Time `json:"sessionDate"`
	Title       string    `json:"title"`
}

func SessionResponseFromSession(session *models.Session) SessionResponse {
	return SessionResponse{
		ID:          session.ID,
		Title:       session.Title,
		SessionDate: session.SessionDate,
	}
}

type TranscriptResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

func TranscriptResponseFromTranscript(transcript *models.Transcript) TranscriptResponse {
	return TranscriptResponse{
		ID:     transcript.JobID,
		Status: transcript.Status.String(),
	}
}

type userManager interface {
	AddNewUser(ctx context.Context, user models.User) (*models.User, error)
	GetUserByID(ctx context.Context, email string) (*models.User, error)
}

type campaignManager interface {
	AddCampaign(ctx context.Context, ownerID string, campaign models.Campaign) (*models.Campaign, error)
	GetCampaignsForUser(ctx context.Context, ownerID string) ([]models.Campaign, error)
}

type sessionManager interface {
	AddSession(ctx context.Context, campaignID string, session models.Session) (*models.Session, error)
	GetSessionsForCampaign(ctx context.Context, campaignID string) ([]models.Session, error)
}

type transcriptionManager interface {
	SubmitTranscriptionJob(ctx context.Context, userID, campaignID, sessionID string, audioFormat models.AudioFormat, audioFile io.Reader) (*models.Transcript, error)
	GetTranscriptJob(ctx context.Context, jobID string) (*models.Transcript, error)
	GetTranscriptsForSession(ctx context.Context, sessionID string) ([]models.Transcript, error)
	DownloadTranscript(ctx context.Context, jobID string, w io.WriterAt) (int64, error)
}

var (
	baseUrl             = "dragonspeak-service"
	maxFileDownloadSize = 10 * 1024 * 1024
)

type HttpAPI struct {
	userManager          userManager
	campaignManager      campaignManager
	sessionManager       sessionManager
	transcriptionManager transcriptionManager
	engine               *gin.Engine
}

func NewHttpAPI(engine *gin.Engine, userManager userManager, campaignManager campaignManager, sessionManager sessionManager, transcriptionManager transcriptionManager) *HttpAPI {
	api := &HttpAPI{
		engine:               engine,
		userManager:          userManager,
		campaignManager:      campaignManager,
		sessionManager:       sessionManager,
		transcriptionManager: transcriptionManager,
	}
	api.registerHandlers()

	return api
}

func (api *HttpAPI) Run() {
	api.engine.Run(":8080")
}

func (api *HttpAPI) registerHandlers() {
	api.engine.POST(baseUrl+"/v1/users", api.AddUser)
	api.engine.GET(baseUrl+"/v1/users/:userId", api.GetUserByID)
	api.engine.POST(baseUrl+"/v1/users/:userId/campaigns", api.AddCampaign)
	api.engine.GET(baseUrl+"/v1/users/:userId/campaigns", api.GetCampaigns)
	api.engine.POST(baseUrl+"/v1/users/:userId/campaigns/:campaignId/sessions", api.AddSession)
	api.engine.GET(baseUrl+"/v1/users/:userId/campaigns/:campaignId/sessions", api.GetSessions)
	api.engine.POST(baseUrl+"/v1/users/:userId/campaigns/:campaignId/sessions/:sessionId/transcripts", api.SubmitTranscriptionJob)
	api.engine.GET(baseUrl+"/v1/users/:userId/campaigns/:campaignId/sessions/:sessionId/transcripts", api.GetTranscriptJobs)
	api.engine.GET(baseUrl+"/v1/users/:userId/campaigns/:campaignId/sessions/:sessionId/transcripts/:jobId", api.GetTranscriptJob)
	api.engine.GET(baseUrl+"/v1/users/:userId/campaigns/:campaignId/sessions/:sessionId/transcripts/:jobId/fulltext", api.GetTranscriptFullText)
}

func (api *HttpAPI) AddUser(c *gin.Context) {
	var user CreateUserRequest
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
		return
	}
	addedUser, err := api.userManager.AddNewUser(c.Request.Context(), user.toUser())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, UserResponseFromUser(addedUser))
}

func (api *HttpAPI) GetUserByID(c *gin.Context) {
	userID := c.Param("userId")
	user, err := api.userManager.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, UserResponseFromUser(user))
}

func (api *HttpAPI) AddCampaign(c *gin.Context) {
	userID := c.Param("userId")
	var campaign CreateCampaignRequest
	err := json.NewDecoder(c.Request.Body).Decode(&campaign)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
		return
	}
	addedCampaign, err := api.campaignManager.AddCampaign(c.Request.Context(), userID, campaign.toCampaign())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, CampaignResponseFromCampaign(addedCampaign))
}

func (api *HttpAPI) GetCampaigns(c *gin.Context) {
	userID := c.Param("userId")
	campaigns, err := api.campaignManager.GetCampaignsForUser(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response := []CampaignResponse{}
	for _, campaign := range campaigns {
		response = append(response, CampaignResponseFromCampaign(&campaign))
	}
	c.JSON(http.StatusOK, response)
}

func (api *HttpAPI) AddSession(c *gin.Context) {
	campaignID := c.Param("campaignId")
	var session CreateSessionRequest
	err := json.NewDecoder(c.Request.Body).Decode(&session)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
	}
	addedSession, err := api.sessionManager.AddSession(c.Request.Context(), campaignID, session.toSession())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, SessionResponseFromSession(addedSession))
}

func (api *HttpAPI) GetSessions(c *gin.Context) {
	campaignID := c.Param("campaignId")
	sessions, err := api.sessionManager.GetSessionsForCampaign(c.Request.Context(), campaignID)
	if err != nil {
		handleError(c, err)
		return
	}
	response := []SessionResponse{}
	for _, session := range sessions {
		response = append(response, SessionResponseFromSession(&session))
	}
	c.JSON(http.StatusOK, response)
}

func (api *HttpAPI) SubmitTranscriptionJob(c *gin.Context) {
	userID := c.Param("userId")
	campaignID := c.Param("campaignId")
	sessionID := c.Param("sessionId")
	fileType := c.Request.Header.Get("Content-Type")
	audioFormat, err := contentTypeToAudioType(fileType)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Unprocessable Entity. Content-Type: %s not supported. Supported types are \"audio/mpeg\", \"audio/webm\", \"audio/ogg\"",
		})
		return
	}
	job, err := api.transcriptionManager.SubmitTranscriptionJob(c.Request.Context(), userID, campaignID, sessionID, audioFormat, c.Request.Body)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, TranscriptResponseFromTranscript(job))
}

func (api *HttpAPI) GetTranscriptJob(c *gin.Context) {
	jobID := c.Param("jobId")
	job, err := api.transcriptionManager.GetTranscriptJob(c.Request.Context(), jobID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, TranscriptResponseFromTranscript(job))
}

func (api *HttpAPI) GetTranscriptJobs(c *gin.Context) {
	sessionID := c.Param("sessionId")
	transcripts, err := api.transcriptionManager.GetTranscriptsForSession(c.Request.Context(), sessionID)
	if err != nil {
		handleError(c, err)
		return
	}
	response := []TranscriptResponse{}
	for _, transcript := range transcripts {
		response = append(response, TranscriptResponseFromTranscript(&transcript))
	}
	c.JSON(http.StatusOK, response)
}

func (api *HttpAPI) GetTranscriptFullText(c *gin.Context) {
	jobID := c.Param("jobId")

	buf := make([]byte, 100)
	writeBuffer := aws.NewWriteAtBuffer(buf)
	bytesWritten, err := api.transcriptionManager.DownloadTranscript(c.Request.Context(), jobID, writeBuffer)
	if err != nil {
		handleError(c, err)
		return
	}

	c.String(http.StatusOK, string(writeBuffer.Bytes()[:bytesWritten]))
}

func contentTypeToAudioType(contentType string) (models.AudioFormat, error) {
	switch contentType {
	case "audio/mpeg":
		return models.MP3, nil
	case "audio/webm":
		return models.WebM, nil
	case "audio/ogg":
		return models.OGG, nil
	}
	return 0, models.InvalidEntity
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, models.EntityNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			ErrorMessage: "Not Found",
		})
	} else if errors.Is(err, models.Conflicted) {
		c.JSON(http.StatusConflict, ErrorResponse{
			ErrorMessage: "Conflict",
		})
	} else if errors.Is(err, models.EntityAlreadyExists) {
		c.JSON(http.StatusConflict, ErrorResponse{
			ErrorMessage: "Already Exists",
		})
	} else if errors.Is(err, models.InvalidEntity) {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Invalid Request",
		})
	} else {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: "Internal Server Error",
		})
	}
}
