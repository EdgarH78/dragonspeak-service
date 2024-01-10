package app

import (
	"context"
	"fmt"

	"github.com/EdgarH78/dragonspeak-service/models"
)

type userDb interface {
	AddNewUser(ctx context.Context, user models.User) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserManager struct {
	userDb userDb
}

func NewUserManager(userDb userDb) *UserManager {
	return &UserManager{
		userDb: userDb,
	}
}

func (u *UserManager) AddNewUser(ctx context.Context, user models.User) (*models.User, error) {
	if user.Email == "" {
		return nil, fmt.Errorf("missing field: Email %w", models.InvalidEntity)
	}
	if user.Handle == "" {
		return nil, fmt.Errorf("missing field: Handle %w", models.InvalidEntity)
	}
	return u.userDb.AddNewUser(ctx, user)
}

func (u *UserManager) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return u.userDb.GetUserByEmail(ctx, email)
}
