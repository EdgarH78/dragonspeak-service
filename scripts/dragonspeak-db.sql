use dragonspeak-db;

CREATE TABLE Users(
    UserKey SERIAL PRIMARY KEY,
    UserId VARCHAR(32) NOT NULL,
    Handle VARCHAR(24) NOT NULL,
    Email VARCHAR(64) NOT NULL
);
CREATE UNIQUE INDEX users_idx_userid ON Users(UserId);
CREATE UNIQUE INDEX users_idx_email ON Users(Email);


CREATE TABLE Campaigns(
    CampaignKey SERIAL PRIMARY KEY,
    CampaignId VARCHAR(32) NOT NULL,
    OwnerUserId INT NOT NULL,
    CampaignName VARCHAR(24) NOT NULL,
    CampaignLink VARCHAR(255) NULL,
    FOREIGN KEY (OwnerUserId) REFERENCES Users(UserKey)
);
CREATE UNIQUE INDEX campaigns_idx_campaignid ON Campaigns(CampaignId);

CREATE TABLE PlayerType(
    PlayerType VARCHAR(16) PRIMARY KEY NOT NULL
);

INSERT INTO PlayerType(PlayerType)
VALUES ('StandardPlayer'),
       ('GM');

CREATE TABLE Players(
    PlayerKey SERIAL PRIMARY KEY,
    PlayerID VARCHAR(32) NOT NULL,
    CampaignKey INT NOT NULL,
    PlayerName VARCHAR(24),
    PlayerType VARCHAR(16) NOT NULL,
    FOREIGN KEY (CampaignKey) REFERENCES Campaigns(CampaignKey),
    FOREIGN KEY (PlayerType) REFERENCES PlayerType(PlayerType)
);
CREATE UNIQUE INDEX players_idx_playerId ON Players(PlayerID);
CREATE UNIQUE INDEX players_idx_playerName_campaignKey ON Players(PlayerName, CampaignKey);

CREATE TABLE Characters(
    CharacterKey SERIAL PRIMARY KEY,
    PlayerKey INT NOT NULL,
    CharacterName VARCHAR(24) NOT NULL,
    CharacterLink VARCHAR(255) NULL,
    FOREIGN KEY (PlayerKey) REFERENCES Players(PlayerKey)
);

CREATE TABLE Sessions(
    SessionKey SERIAL PRIMARY KEY,
    SessionId VARCHAR(32) NOT NULL,
    CampaignKey INT NOT NULL,
    SessionDate DATE NOT NULL,
    Title VARCHAR(24) NULL,  
    FOREIGN KEY (CampaignKey) REFERENCES Campaigns(CampaignKey)  
);
CREATE UNIQUE INDEX sessions_idx_sessionId ON Sessions(SessionId);
CREATE INDEX sessions_idx_campaignkey_sessiondate  ON Sessions(CampaignKey, SessionDate);

CREATE TABLE SessionAttendance(
    SessionKey INT,
    PlayerKey INT,
    FOREIGN KEY(SessionKey) REFERENCES Sessions(SessionKey),
    FOREIGN KEY (PlayerKey) REFERENCES Players(PlayerKey)
);

CREATE TABLE TranscriptionStatus(
    Status VARCHAR(32) PRIMARY KEY
);

INSERT INTO TranscriptionStatus(Status)
VALUES ('NotStarted'),
       ('Transcribing'),
       ('Summarizing'),
       ('Done'),
       ('TranscriptionFailed'),
       ('SummarizingFailed');

CREATE TABLE SessionTranscripts(
    TranscriptKey SERIAL PRIMARY KEY,
    SessionId INT NOT NULL,
    TranscriptionJobId VARCHAR(128) NULL,
    AudioLocation VARCHAR(128) NULL,
    AudioFormat VARCHAR(10) NULL,
    TranscriptLocation VARCHAR(128) NULL,
    SummaryLocation VARCHAR(128) NULL,
    Status VARCHAR(16) NOT NULL,
    FOREIGN KEY (Status) REFERENCES TranscriptionStatus(Status)
);
CREATE INDEX sessiontranscripts_idx_status ON SessionTranscripts(Status);

