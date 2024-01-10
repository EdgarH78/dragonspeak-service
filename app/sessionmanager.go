package app

import (
	"context"
	"fmt"

	"github.com/EdgarH78/dragonspeak-service/models"
)

type sessionDb interface {
	AddSession(ctx context.Context, campaignID string, session models.Session) (*models.Session, error)
	GetSessionsForCampaign(ctx context.Context, campaignID string) ([]models.Session, error)
}

type SessionManager struct {
	sessionDb sessionDb
}

func NewSessionManager(sessionDb sessionDb) *SessionManager {
	return &SessionManager{
		sessionDb: sessionDb,
	}
}

func (s *SessionManager) AddSession(ctx context.Context, campaignID string, session models.Session) (*models.Session, error) {
	if session.Title == "" {
		return nil, fmt.Errorf("missing field: Title %w", models.InvalidEntity)
	}
	return s.sessionDb.AddSession(ctx, campaignID, session)
}

func (s *SessionManager) GetSessionsForCampaign(ctx context.Context, campaignID string) ([]models.Session, error) {
	return s.sessionDb.GetSessionsForCampaign(ctx, campaignID)
}
