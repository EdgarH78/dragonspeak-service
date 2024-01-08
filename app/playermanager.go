package app

import "github.com/EdgarH78/dragonspeak-service/models"

type playerDB interface {
	AddNewPlayer(campaignID string, player models.Player) (*models.Player, error)
	GetPlayersForCampaign(campaignID string) ([]models.Player, error)
}

type PlayerManager struct {
	playerDB playerDB
}

func NewPlayerManager(playerDB playerDB) *PlayerManager {
	return &PlayerManager{
		playerDB: playerDB,
	}
}
