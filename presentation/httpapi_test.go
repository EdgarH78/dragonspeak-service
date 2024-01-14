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

func (m *MockUserManager) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
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
				var actualUserResponse UserResponse
				err = json.Unmarshal(w.Body.Bytes(), &actualUserResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}

				assert.Equal(t, c.expectedUserReponse.Email, actualUserResponse.Email)
				assert.Equal(t, c.expectedUserReponse.Handle, actualUserResponse.Handle)
				assert.Equal(t, c.expectedUserReponse.ID, actualUserResponse.ID)
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

func TestGetUserByEmail(t *testing.T) {
	cases := []struct {
		description           string
		userID                string
		managerUserResponse   *models.User
		managerError          error
		expectedUserReponse   *UserResponse
		expectedErrorResponse *ErrorResponse
		expectedStatusCode    int
	}{
		{
			description: "user returned",
			userID:      "testUID",
			managerUserResponse: &models.User{
				ID:     "testUID",
				Handle: "testuser",
				Email:  "test@test.com",
			},
			expectedUserReponse: &UserResponse{
				Handle: "testuser",
				Email:  "test@test.com",
				ID:     "testUID",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			description:        "user not found",
			userID:             "testUID",
			managerError:       models.EntityNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Not Found",
			},
		},
		{
			description:        "internal server error",
			userID:             "testUID",
			managerError:       fmt.Errorf("could not connect to database"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Internal Server Error",
			},
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
				userManager.On("GetUserByID", mock.Anything, c.userID).Return(c.managerUserResponse, nil)
			} else if c.managerError != nil {
				userManager.On("GetUserByID", mock.Anything, mock.Anything).Return(nil, c.managerError)
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/dragonspeak-service/v1/users/%s", c.userID), nil)
			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedUserReponse != nil {
				var actualUserResponse UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualUserResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}

				assert.Equal(t, c.expectedUserReponse.Email, actualUserResponse.Email)
				assert.Equal(t, c.expectedUserReponse.Handle, actualUserResponse.Handle)
				assert.Equal(t, c.expectedUserReponse.ID, actualUserResponse.ID)
			} else if c.expectedErrorResponse != nil {
				var actualErrorResponse ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualErrorResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}
				assert.Equal(t, c.expectedErrorResponse.ErrorMessage, actualErrorResponse.ErrorMessage)
			}
		})
	}
}

func TestCreateCampaign(t *testing.T) {
	cases := []struct {
		description              string
		userID                   string
		campaign                 *models.Campaign
		managerCampaignResponse  *models.Campaign
		managerError             error
		expectedCampaignResponse *CampaignResponse
		expectedErrorResponse    *ErrorResponse
		expectedStatusCode       int
	}{
		{
			description: "campaign created",
			userID:      "abc123",
			campaign: &models.Campaign{
				ID:   "cmp123",
				Name: "one-shot",
				Link: "http://abc.com",
			},
			managerCampaignResponse: &models.Campaign{
				ID:   "cmp123",
				Name: "one-shot",
				Link: "http://abc.com",
			},
			expectedCampaignResponse: &CampaignResponse{
				ID:   "cmp123",
				Name: "one-shot",
				Link: "http://abc.com",
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			description: "database error returned",
			userID:      "abc123",
			campaign: &models.Campaign{
				ID:   "cmp123",
				Name: "one-shot",
				Link: "http://abc.com",
			},
			managerError: fmt.Errorf("something went wrong"),
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Internal Server Error",
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			description: "conflicted error returned",
			userID:      "abc123",
			campaign: &models.Campaign{
				ID:   "cmp123",
				Name: "one-shot",
				Link: "http://abc.com",
			},
			managerError: models.Conflicted,
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Conflict",
			},
			expectedStatusCode: http.StatusConflict,
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
			if c.expectedCampaignResponse != nil {
				campaignManager.On("AddCampaign", mock.Anything, c.userID, mock.Anything).Return(c.managerCampaignResponse, nil)
			} else if c.managerError != nil {
				campaignManager.On("AddCampaign", mock.Anything, mock.Anything, mock.Anything).Return(nil, c.managerError)
			}

			campaignBody, err := json.Marshal(c.campaign)
			if err != nil {
				t.Errorf("unexpected error marshaling cage to json: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("POST", fmt.Sprintf("/dragonspeak-service/v1/users/%s/campaigns", c.userID), bytes.NewReader(campaignBody))
			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedCampaignResponse != nil {
				var actualCampaignResponse CampaignResponse
				err = json.Unmarshal(w.Body.Bytes(), &actualCampaignResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}

				assert.Equal(t, c.expectedCampaignResponse.Name, actualCampaignResponse.Name)
				assert.Equal(t, c.expectedCampaignResponse.Link, actualCampaignResponse.Link)
				assert.Equal(t, c.expectedCampaignResponse.ID, actualCampaignResponse.ID)
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

func TestGetCampaigns(t *testing.T) {
	cases := []struct {
		description               string
		userID                    string
		managerCampaignsResponse  []models.Campaign
		managerError              error
		expectedCampaignsResponse []CampaignResponse
		expectedErrorResponse     *ErrorResponse
		expectedStatusCode        int
	}{
		{
			description: "campaigns returned",
			userID:      "abc123",
			managerCampaignsResponse: []models.Campaign{
				{
					ID:   "cmp123",
					Name: "one-shot",
					Link: "http://campaign1",
				},
				{
					ID:   "cmp456",
					Name: "two-shot",
					Link: "http://campaign2",
				},
			},
			expectedCampaignsResponse: []CampaignResponse{
				{
					ID:   "cmp123",
					Name: "one-shot",
					Link: "http://campaign1",
				},
				{
					ID:   "cmp456",
					Name: "two-shot",
					Link: "http://campaign2",
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			description:        "campaigns not found",
			userID:             "abc123",
			managerError:       models.EntityNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Not Found",
			},
		},
		{
			description:        "internal server error",
			userID:             "abc123",
			managerError:       fmt.Errorf("an error has occurred"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorResponse: &ErrorResponse{
				ErrorMessage: "Internal Server Error",
			},
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
			if c.expectedCampaignsResponse != nil {
				campaignManager.On("GetCampaignsForUser", mock.Anything, c.userID).Return(c.managerCampaignsResponse, nil)
			} else if c.managerError != nil {
				campaignManager.On("GetCampaignsForUser", mock.Anything, mock.Anything).Return(nil, c.managerError)
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/dragonspeak-service/v1/users/%s/campaigns", c.userID), nil)
			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
				return
			}
			if c.expectedCampaignsResponse != nil {
				var actualCampaignsResponse []CampaignResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualCampaignsResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}
				if len(c.expectedCampaignsResponse) != len(actualCampaignsResponse) {
					t.Errorf("expected %d campaigns got %d", len(c.expectedCampaignsResponse), len(actualCampaignsResponse))
					return
				}

				for i, expected := range c.expectedCampaignsResponse {
					actual := actualCampaignsResponse[i]
					assert.Equal(t, expected.Name, actual.Name)
					assert.Equal(t, expected.Link, actual.Link)
					assert.Equal(t, expected.ID, actual.ID)
				}

			} else if c.expectedErrorResponse != nil {
				var actualErrorResponse ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &actualErrorResponse)
				if err != nil {
					t.Fatalf("unexpected error when unmarshalling response: %s", err)
					return
				}
				assert.Equal(t, c.expectedErrorResponse.ErrorMessage, actualErrorResponse.ErrorMessage)
			}
		})
	}
}
