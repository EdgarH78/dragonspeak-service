package database

import (
	"database/sql"
	"fmt"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	maxOpenConns = 20
	maxIdleConns = 10
)

type SQLConfig struct {
	User         string
	Password     string
	Host         string
	Port         string
	DatabaseName string
}

func (s SQLConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", s.Host, s.Port, s.User, s.Password, s.DatabaseName)
}

type PostgresDao struct {
	db *sql.DB
}

// NewPostgresDao creates a new instance of PostgresDao
func NewPostgresDao(config SQLConfig) (*PostgresDao, error) {
	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	return &PostgresDao{db: db}, nil
}

// AddNewUser adds a new user to the Users table
func (dao *PostgresDao) AddNewUser(user models.User) (*models.User, error) {
	userID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	_, err = dao.db.Exec("INSERT INTO Users (UserId, Handle, Email) VALUES ($1, $2, $3)", userID.String(), user.Handle, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:     userID.String(),
		Handle: user.Handle,
	}, nil
}

func (dao *PostgresDao) GetUserByEmail(email string) (*models.User, error) {
	qs := `SELECT UserId, Handle, Email
			FROM Users
			WHERE Email = $1`
	rows, err := dao.db.Query(qs, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, models.EntityNotFound
	}

	user := models.User{}
	if err = rows.Scan(&user.ID, &user.Handle, &user.Email); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateCampaign creates a new campaign
func (dao *PostgresDao) CreateCampaign(ownerID string, campaign models.Campaign) (*models.Campaign, error) {
	campaignID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	insertStmt := `INSERT INTO Campaigns (CampaignId, OwnerUserId, CampaignName, CampaignLink) 
					SELECT $1, UserKey, $2, $3 
					FROM Users
					WHERE UserId=$4`

	_, err = dao.db.Exec(insertStmt, campaignID.String(), campaign.Name, campaign.Link, ownerID)
	if err != nil {
		return nil, err
	}
	return &models.Campaign{
		ID:   campaignID.String(),
		Name: campaign.Name,
		Link: campaign.Link,
	}, nil
}

// GetAllCampaignsForUser retrieves all campaigns for a specific user
func (dao *PostgresDao) GetCampaignsForUser(ownerID string) ([]models.Campaign, error) {
	qs := ` SELECT c.CampaignId, c.CampaignName, c.CampaignLink
			FROM Campaign c
			JOIN Users u on u.UserKey=c.OwnerUserId
			WHERE u.UserId=$1
	`
	rows, err := dao.db.Query(qs, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	campaigns := []models.Campaign{}
	for rows.Next() {
		c := models.Campaign{}

		if err := rows.Scan(&c.ID, &c.Name, &c.Link); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

// AddNewPlayer adds a new player to the Players table
func (dao *PostgresDao) AddNewPlayer(campaignID string, player models.Player) (*models.Player, error) {
	playerID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	insertStmt := `INSERT INTO Players(CampaignKey, PlayerID, PlayerName, PlayerType) 
					SELECT CampaignKey, $1, $2, $3 
					FROM Campaigns
					WHERE CampaignId=$4`
	_, err = dao.db.Exec(insertStmt, playerID.String(), player.Name, player.Type.String(), campaignID)
	if err != nil {
		return nil, err
	}
	return &models.Player{
		ID:   playerID.String(),
		Name: player.Name,
		Type: player.Type,
	}, nil
}

func (dao *PostgresDao) GetPlayersForCampaign(campaignID string) ([]models.Player, error) {
	qs := `SELECT p.PlayerID, p.PlayerName, p.PlayerType 
		   FROM Players p 
		   JOIN Campaign c on c.CampaignKey=p.CampaignKey 
		   WHERE c.CampaignId = $1`

	rows, err := dao.db.Query(qs, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		p := models.Player{}
		playerTypeStr := ""
		if err := rows.Scan(&p.ID, &p.Name, &playerTypeStr); err != nil {
			return nil, err
		}
		var playerType models.PlayerType
		if playerType, err = models.PlayerTypeFromString(playerTypeStr); err != nil {
			return nil, err
		}
		p.Type = playerType
		players = append(players, p)
	}
	return players, nil
}

func (dao *PostgresDao) CreateSession(campaignID string, session models.Session) (*models.Session, error) {
	sessionID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	insertStmt := `INSERT INTO Sessions(SessionId, CampaignKey, SessionDate, Title) 
				  SELECT $1, CampaignKey, $2, $3 
				  FROM Campaigns WHERE CampaignId=$4`
	_, err = dao.db.Exec(insertStmt, sessionID.String(), session.SessionDate, session.Title, campaignID)
	if err != nil {
		return nil, err
	}
	return &models.Session{
		ID:          sessionID.String(),
		SessionDate: session.SessionDate,
		Title:       session.Title,
	}, nil
}

func (dao *PostgresDao) GetSessionsForCampaign(campaignID string) ([]models.Session, error) {
	qs := `SELECT s.SessionId, s.SessionDate, s.Title
		   FROM Sessions s 
		   JOIN Campaign c ON c.CampaignKey = s.CampaignKey 
		   WHERE c.CampaignId = $1`

	rows, err := dao.db.Query(qs, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []models.Session{}
	for rows.Next() {
		session := models.Session{}
		if err = rows.Scan(&session.ID, &session.SessionDate, &session.Title); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (dao *PostgresDao) AddTranscriptToSessions(sessionID string, transcript models.Transcript) (*models.Transcript, error) {
	insertStmt := `INSERT INTO SessionTranscripts(SessionId, TranscriptionJobId, AudioLocation, AudioFormat, TranscriptLocation, SummaryLocation, Status)
				   SELECT SessionKey, $1, $2, $3, $4, $5, $6
				   FROM Sessions 
				   WHERE SessionId=$7`
	_, err := dao.db.Exec(insertStmt, transcript.JobID, transcript.AudioLocation, transcript.AudioFormat, transcript.TranscriptLocation, transcript.SummaryLocation, transcript.Status.String())
	if err != nil {
		return nil, err
	}
	return &transcript, nil
}

func (dao *PostgresDao) GetTranscriptsForSessions(sessionID string) ([]models.Transcript, error) {
	qs := `SELECT t.TranscriptionJobId, t.AudioLocation, t.AudioFormat, t.TranscriptLocation, t.SummaryLocation, t.Status 
		   FROM SessionTranscripts t 
		   JOIN Sessions s on s.SessionKey = t.SessionId 
		   WHERE s.SessionId=$1`
	rows, err := dao.db.Query(qs, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transcripts := []models.Transcript{}
	for rows.Next() {
		transcript := models.Transcript{}
		statusString := ""
		if err = rows.Scan(&transcript.JobID, &transcript.AudioLocation, &transcript.TranscriptLocation, &transcript.SummaryLocation, &statusString); err != nil {
			return nil, err
		}
		transcriptStatus, err := models.TranscriptStatusFromString(statusString)
		if err != nil {
			return nil, err
		}
		transcript.Status = transcriptStatus
		transcripts = append(transcripts, transcript)
	}
	return transcripts, nil
}
