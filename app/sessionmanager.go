package app

import (
	"fmt"

	"github.com/EdgarH78/dragonspeak-service/models"
)

type sessionDb interface {
	AddSession(campaignID string, session models.Session) (*models.Session, error)
	GetSessionsForCampaign(campaignID string) ([]models.Session, error)
}

type SessionManager struct {
	sessionDb sessionDb
}

func NewSessionManager(sessionDb sessionDb) *SessionManager {
	return &SessionManager{
		sessionDb: sessionDb,
	}
}

func (s *SessionManager) AddSession(campaignID string, session models.Session) (*models.Session, error) {
	if session.Title == "" {
		return nil, fmt.Errorf("missing field: Title %w", models.InvalidEntity)
	}
	return s.sessionDb.AddSession(campaignID, session)
}

func (s *SessionManager) GetSessionsForCampaign(campaignID string) ([]models.Session, error) {
	return s.sessionDb.GetSessionsForCampaign(campaignID)
}
