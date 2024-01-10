package app

import (
	"context"
	"fmt"

	"github.com/EdgarH78/dragonspeak-service/models"
)

type campaignDb interface {
	AddCampaign(ctx context.Context, ownerID string, campaign models.Campaign) (*models.Campaign, error)
	GetCampaignsForUser(ctx context.Context, ownerID string) ([]models.Campaign, error)
}

type CampaignManager struct {
	campaignDb campaignDb
}

func NewCampaignManager(campaignDb campaignDb) *CampaignManager {
	return &CampaignManager{
		campaignDb: campaignDb,
	}
}

func (c *CampaignManager) AddCampaign(ctx context.Context, ownerID string, campaign models.Campaign) (*models.Campaign, error) {
	if campaign.Name == "" {
		return nil, fmt.Errorf("missing field name %w", models.InvalidEntity)
	}
	return c.campaignDb.AddCampaign(ctx, ownerID, campaign)
}

func (c *CampaignManager) GetCampaignsForUser(ctx context.Context, ownerID string) ([]models.Campaign, error) {
	return c.campaignDb.GetCampaignsForUser(ctx, ownerID)
}
