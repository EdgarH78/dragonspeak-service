package app

import (
	"context"
	"errors"
	"testing"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserDb struct {
	mock.Mock
}

func (m *MockUserDb) AddNewUser(ctx context.Context, user models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDb) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), nil
}

func TestAddNewUser(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		userToAdd      models.User
		dbError        error
		dbResult       *models.User
		expectedError  error
		expectedResult *models.User
	}{
		{
			description: "user is added to the database",
			userToAdd: models.User{
				Handle: "testHandle",
				Email:  "test@test.com",
			},
			dbResult: &models.User{
				Handle: "testHandle",
				Email:  "test@test.com",
				ID:     "abc123",
			},
			expectedResult: &models.User{
				Handle: "testHandle",
				Email:  "test@test.com",
				ID:     "abc123",
			},
		},
		{
			description: "user does not have a handle, InvalidEntity returned",
			userToAdd: models.User{
				Email: "test@test.com",
			},
			expectedError: models.InvalidEntity,
		},
		{
			description: "user does not have an email, InvalidEntity returned",
			userToAdd: models.User{
				Handle: "testHandle",
			},
			expectedError: models.InvalidEntity,
		},
		{
			description: "database returned an error, error is returned",
			userToAdd: models.User{
				Handle: "testHandle",
				Email:  "test@test.com",
			},
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockUserDb{}
			if c.dbError != nil {
				mockDb.On("AddNewUser", mock.Anything, c.userToAdd).Return(nil, c.dbError)
			} else {
				mockDb.On("AddNewUser", mock.Anything, c.userToAdd).Return(c.dbResult, nil)
			}
			testManager := NewUserManager(mockDb)
			result, err := testManager.AddNewUser(context.Background(), c.userToAdd)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				assert.Equal(t, c.expectedResult.Handle, result.Handle)
				assert.Equal(t, c.expectedResult.Email, result.Email)
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

func TestGetUserByEmail(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		email          string
		dbError        error
		dbResult       *models.User
		expectedError  error
		expectedResult *models.User
	}{
		{
			description: "user is retrieved",
			email:       "test@test.com",
			dbResult: &models.User{
				ID:     "abc123",
				Handle: "testUser",
				Email:  "test@test.com",
			},
			expectedResult: &models.User{
				ID:     "abc123",
				Handle: "testUser",
				Email:  "test@test.com",
			},
		},
		{
			description:   "database returns error",
			email:         "test@test.com",
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockUserDb{}
			if c.dbError != nil {
				mockDb.On("GetUserByEmail", mock.Anything, c.email).Return(nil, c.dbError)
			} else {
				mockDb.On("GetUserByEmail", mock.Anything, c.email).Return(c.dbResult, nil)
			}
			testManager := NewUserManager(mockDb)
			result, err := testManager.GetUserByEmail(context.Background(), c.email)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				assert.Equal(t, c.expectedResult.Handle, result.Handle)
				assert.Equal(t, c.expectedResult.Email, result.Email)
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
