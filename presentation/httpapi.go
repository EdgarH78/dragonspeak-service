package presentation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/EdgarH78/dragonspeak-service/models"
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
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
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
	DownloadTranscript(ctx context.Context, jobID string, w io.WriterAt) error
}

var (
	baseUrl = "dragonspeak-service"
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
		handlError(c, err)
		return
	}
	c.JSON(http.StatusCreated, UserResponseFromUser(addedUser))
}

func handlError(c *gin.Context, err error) {
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
