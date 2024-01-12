package presentation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserManager struct {
	mock.Mock
}

func (m *MockUserManager) AddNewUser(ctx context.Context, user models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), nil
}

func (m *MockUserManager) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), nil
}

type MockCampaignManager struct {
	mock.Mock
}

func (m *MockCampaignManager) AddCampaign(ctx context.Context, ownerID string, campaign models.Campaign) (*models.Campaign, error) {
	args := m.Called(ctx, ownerID, campaign)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Campaign), nil
}

func (m *MockCampaignManager) GetCampaignsForUser(ctx context.Context, ownerID string) ([]models.Campaign, error) {
	args := m.Called(ctx, ownerID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Campaign), nil
}

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) AddSession(ctx context.Context, campaignID string, session models.Session) (*models.Session, error) {
	args := m.Called(ctx, campaignID, session)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), nil
}

func (m *MockSessionManager) GetSessionsForCampaign(ctx context.Context, campaignID string) ([]models.Session, error) {
	args := m.Called(ctx, campaignID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), nil
}

type MockTranscriptionManager struct {
	mock.Mock
}

func (m *MockTranscriptionManager) SubmitTranscriptionJob(ctx context.Context, userID, campaignID, sessionID string, audioFormat models.AudioFormat, audioFile io.Reader) (*models.Transcript, error) {
	args := m.Called(ctx, userID, campaignID, sessionID, audioFormat, audioFile)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), nil
}

func (m *MockTranscriptionManager) GetTranscriptJob(ctx context.Context, jobID string) (*models.Transcript, error) {
	args := m.Called(ctx, jobID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), nil
}

func (m *MockTranscriptionManager) GetTranscriptsForSession(ctx context.Context, sessionID string) ([]models.Transcript, error) {
	args := m.Called(ctx, sessionID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Transcript), nil
}

func (m *MockTranscriptionManager) DownloadTranscript(ctx context.Context, jobID string, w io.WriterAt) error {
	args := m.Called(ctx, jobID, w)
	return args.Error(0)
}

func TestAddUser(t *testing.T) {
	cases := []struct {
		description           string
		crateUserRequuest     CreateUserRequest
		managerUserResponse   *models.User
		managerError          error
		expectedUserReponse   *UserResponse
		expectedErrorResponse *ErrorResponse
		expectedStatusCode    int
	}{
		{
			description: "user added request, user is added",
			crateUserRequuest: CreateUserRequest{
				Handle: "test",
				Email:  "test@test.com",
			},
			managerUserResponse: &models.User{
				Handle: "test",
				Email:  "test@test.com",
				ID:     "abc123",
			},
			expectedUserReponse: &UserResponse{
				Handle: "test",
				Email:  "test@test.com",
				ID:     "abc123",
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			description: "user added request, EntityAlreadyExist error returned",
			crateUserRequuest: CreateUserRequest{
				Handle: "test",
				Email:  "test@test.com",
			},
			managerError: models.EntityAlreadyExists,
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Already Exists",
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			description: "user added request, database error error returned",
			crateUserRequuest: CreateUserRequest{
				Handle: "test",
				Email:  "test@test.com",
			},
			managerError: fmt.Errorf("failed to connect to database host: abc"),
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Internal Server Error",
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			userManager := &MockUserManager{}
			campaignManager := &MockCampaignManager{}
			sessionManager := &MockSessionManager{}
			transcriptionManager := &MockTranscriptionManager{}

			NewHttpAPI(r, userManager, campaignManager, sessionManager, transcriptionManager)
			if c.managerUserResponse != nil {
				userManager.On("AddNewUser", mock.Anything, mock.Anything).Return(c.managerUserResponse, nil)
			} else if c.managerError != nil {
				userManager.On("AddNewUser", mock.Anything, mock.Anything).Return(nil, c.managerError)
			}

			userBody, err := json.Marshal(c.crateUserRequuest)
			if err != nil {
				t.Errorf("unexpected error marshaling cage to json: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("POST", "/dragonspeak-service/v1/users", bytes.NewReader(userBody))
			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedUserReponse != nil {
				var actualUserRespones UserResponse
				err = json.Unmarshal(w.Body.Bytes(), &actualUserRespones)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}

				assert.Equal(t, c.expectedUserReponse.Email, actualUserRespones.Email)
				assert.Equal(t, c.expectedUserReponse.Handle, actualUserRespones.Handle)
				assert.Equal(t, c.expectedUserReponse.ID, actualUserRespones.ID)
			} else if c.expectedErrorResponse != nil {
				var actualErrorResponse ErrorResponse
				err = json.Unmarshal(w.Body.Bytes(), &actualErrorResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}
				assert.Equal(t, c.expectedErrorResponse.ErrorMessage, actualErrorResponse.ErrorMessage)
			}
		})
	}
}
