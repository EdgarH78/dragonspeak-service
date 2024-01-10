package app

import (
	"context"
	"errors"
	"testing"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCampaignDB struct {
	mock.Mock
}

func (m *MockCampaignDB) AddCampaign(ctx context.Context, ownerID string, campaign models.Campaign) (*models.Campaign, error) {
	args := m.Called(ctx, ownerID, campaign)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Campaign), nil
}

func (m *MockCampaignDB) GetCampaignsForUser(ctx context.Context, ownerID string) ([]models.Campaign, error) {
	args := m.Called(ctx, ownerID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Campaign), nil

}

func TestAddCampaign(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		campaignToAdd  models.Campaign
		userID         string
		dbError        error
		dbResult       *models.Campaign
		expectedError  error
		expectedResult *models.Campaign
	}{
		{
			description: "campaign is added to the database",
			userID:      "userId123",
			campaignToAdd: models.Campaign{
				Name: "testAndDragons",
				Link: "http://dnd.com",
			},
			dbResult: &models.Campaign{
				Name: "testAndDragons",
				Link: "http://dnd.com",
				ID:   "abc123",
			},
			expectedResult: &models.Campaign{
				Name: "testAndDragons",
				Link: "http://dnd.com",
				ID:   "abc123",
			},
		},
		{
			description: "campaign does not have a name, InvalidEntity returned",
			userID:      "userId123",
			campaignToAdd: models.Campaign{
				Link: "http://dnd.com",
			},
			expectedError: models.InvalidEntity,
		},
		{
			description: "database returned an error, error is returned",
			userID:      "userId123",
			campaignToAdd: models.Campaign{
				Name: "testAndDragons",
				Link: "http://dnd.com",
			},
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockCampaignDB{}
			if c.dbError != nil {
				mockDb.On("AddCampaign", mock.Anything, c.userID, c.campaignToAdd).Return(nil, c.dbError)
			} else {
				mockDb.On("AddCampaign", mock.Anything, c.userID, c.campaignToAdd).Return(c.dbResult, nil)
			}
			testManager := NewCampaignManager(mockDb)
			result, err := testManager.AddCampaign(context.Background(), c.userID, c.campaignToAdd)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				assert.Equal(t, c.expectedResult.Link, result.Link)
				assert.Equal(t, c.expectedResult.Name, result.Name)
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

func TestCampaignsForUser(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		userID         string
		dbError        error
		dbResult       []models.Campaign
		expectedError  error
		expectedResult []models.Campaign
	}{
		{
			description: "user is retrieved",
			userID:      "user123",
			dbResult: []models.Campaign{
				{
					Name: "testAndDragons",
					Link: "http://dnd.com",
					ID:   "abc123",
				},
				{
					Name: "testAndDragons2",
					Link: "http://dnd.com",
					ID:   "abc456",
				},
			},
			expectedResult: []models.Campaign{
				{
					Name: "testAndDragons",
					Link: "http://dnd.com",
					ID:   "abc123",
				},
				{
					Name: "testAndDragons2",
					Link: "http://dnd.com",
					ID:   "abc456",
				},
			},
		},
		{
			description:   "database returns error",
			userID:        "abc123",
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockCampaignDB{}
			if c.dbError != nil {
				mockDb.On("GetCampaignsForUser", mock.Anything, c.userID).Return(nil, c.dbError)
			} else {
				mockDb.On("GetCampaignsForUser", mock.Anything, c.userID).Return(c.dbResult, nil)
			}
			testManager := NewCampaignManager(mockDb)
			result, err := testManager.GetCampaignsForUser(context.Background(), c.userID)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				if len(c.expectedResult) != len(result) {
					t.Errorf("expected %d campaigns got %d", len(c.expectedResult), len(result))
					return
				}
				for i, expectedCampaign := range c.expectedResult {
					actualCampaign := result[i]
					assert.Equal(t, expectedCampaign.Link, actualCampaign.Link)
					assert.Equal(t, expectedCampaign.Name, actualCampaign.Name)
					assert.Equal(t, expectedCampaign.ID, actualCampaign.ID)
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
