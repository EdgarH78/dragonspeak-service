package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSessionDB struct {
	mock.Mock
}

func (m *MockSessionDB) AddSession(ctx context.Context, campaignID string, session models.Session) (*models.Session, error) {
	args := m.Called(ctx, campaignID, session)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), nil
}

func (m *MockSessionDB) GetSessionsForCampaign(ctx context.Context, campaignID string) ([]models.Session, error) {
	args := m.Called(ctx, campaignID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), nil
}

func TestAddSession(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		campaignID     string
		sessionToAdd   models.Session
		dbError        error
		dbResult       *models.Session
		expectedError  error
		expectedResult *models.Session
	}{
		{
			description: "session is added to the database",
			campaignID:  "testCampaign123",
			sessionToAdd: models.Session{
				SessionDate: time.Now(),
				Title:       "session-0",
			},
			dbResult: &models.Session{
				SessionDate: time.Now(),
				Title:       "session-0",
				ID:          "abc123",
			},
			expectedResult: &models.Session{
				SessionDate: time.Now(),
				Title:       "session-0",
				ID:          "abc123",
			},
		},
		{
			description: "session does not have a title, InvalidEntity returned",
			sessionToAdd: models.Session{
				SessionDate: time.Now(),
			},
			expectedError: models.InvalidEntity,
		},
		{
			description: "database returned an error, error is returned",
			sessionToAdd: models.Session{
				SessionDate: time.Now(),
				Title:       "session-0",
			},
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockSessionDB{}
			if c.dbError != nil {
				mockDb.On("AddSession", mock.Anything, c.campaignID, c.sessionToAdd).Return(nil, c.dbError)
			} else {
				mockDb.On("AddSession", mock.Anything, c.campaignID, c.sessionToAdd).Return(c.dbResult, nil)
			}
			testManager := NewSessionManager(mockDb)
			result, err := testManager.AddSession(context.Background(), c.campaignID, c.sessionToAdd)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				assert.Equal(t, c.expectedResult.SessionDate, result.SessionDate)
				assert.Equal(t, c.expectedResult.Title, result.Title)
				assert.Equal(t, c.expectedResult.ID, result.ID)
			}
			if c.expectedError != nil {
				if err == nil {
					t.Errorf("expected error: %s got nil", c.expectedError)
					return
				}
				if !errors.Is(err, c.expectedError) {
					t.Errorf("expected error: %s got %s", c.expectedError, err)
				}
			}
		})
	}
}

func TestGetSessionsForCampaign(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		campaignID     string
		dbError        error
		dbResult       []models.Session
		expectedError  error
		expectedResult []models.Session
	}{
		{
			description: "user is retrieved",
			campaignID:  "campaign123",
			dbResult: []models.Session{
				{
					Title:       "session-0",
					SessionDate: time.Now(),
					ID:          "abc123",
				},
				{
					Title:       "session-1",
					SessionDate: time.Now(),
					ID:          "abc456",
				},
			},
			expectedResult: []models.Session{
				{
					Title:       "session-0",
					SessionDate: time.Now(),
					ID:          "abc123",
				},
				{
					Title:       "session-1",
					SessionDate: time.Now(),
					ID:          "abc456",
				},
			},
		},
		{
			description:   "database returns error",
			campaignID:    "campaign123",
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockSessionDB{}
			if c.dbError != nil {
				mockDb.On("GetSessionsForCampaign", mock.Anything, c.campaignID).Return(nil, c.dbError)
			} else {
				mockDb.On("GetSessionsForCampaign", mock.Anything, c.campaignID).Return(c.dbResult, nil)
			}
			testManager := NewSessionManager(mockDb)
			result, err := testManager.GetSessionsForCampaign(context.Background(), c.campaignID)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				if len(c.expectedResult) != len(result) {
					t.Errorf("expected %d campaigns got %d", len(c.expectedResult), len(result))
					return
				}
				for i, expectedSession := range c.expectedResult {
					actualSession := result[i]
					assert.Equal(t, expectedSession.SessionDate, actualSession.SessionDate)
					assert.Equal(t, expectedSession.Title, actualSession.Title)
					assert.Equal(t, expectedSession.ID, actualSession.ID)
				}
			}
			if c.expectedError != nil {
				if err == nil {
					t.Errorf("expected error: %s got nil", c.expectedError)
					return
				}
				if !errors.Is(err, c.expectedError) {
					t.Errorf("expected error: %s got %s", c.expectedError, err)
				}
			}
		})
	}
}
