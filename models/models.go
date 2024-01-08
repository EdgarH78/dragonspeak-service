package models

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID     string
	Handle string
	Email  string
}

// Campaign represents a campaign in the system.
type Campaign struct {
	ID   string
	Name string
	Link string
}

// Character represents a character in the system.
type Character struct {
	OwnerID string
	ID      string
	Name    string
	Link    string
}

// Session represents a session in the system.
type Session struct {
	ID          string
	SessionDate time.Time
	Title       string
}

type PlayerType int

const (
	GM PlayerType = iota
	StandardPlayer
)

// statusStrings maps TranscriptionJobStatusType constants to their string representations
var playerTypeStrings = []string{"GM", "StandardPlayer"}

func (p PlayerType) String() string {
	return playerTypeStrings[p]
}

// FromString converts a string to a TranscriptionJobStatusType
func PlayerTypeFromString(str string) (PlayerType, error) {
	for i, s := range playerTypeStrings {
		if strings.EqualFold(s, str) { // Case insensitive comparison
			return PlayerType(i), nil
		}
	}
	return 0, fmt.Errorf("invalid PlayerType: %s", str)
}

// Player represents a player in the system.
type Player struct {
	ID   string
	Name string
	Type PlayerType // Could be a foreign key to a PlayerType table
}

type TranscriptStatus int

const (
	NotStarted = iota
	Transcribing
	Summarizing
	Done
	TranscriptionFailed
	SummarizingFailed
)

var transcriptStatusStrings = []string{"NotStarted", "Transcribing", "Summarizing", "Done", "TranscriptionFailed", "SummarizingFailed"}

func (t TranscriptStatus) String() string {
	return transcriptStatusStrings[t]
}

func TranscriptStatusFromString(str string) (TranscriptStatus, error) {
	for i, s := range transcriptStatusStrings {
		if strings.EqualFold(s, str) { // Case insensitive comparison
			return TranscriptStatus(i), nil
		}
	}
	return 0, fmt.Errorf("invalid TranscriptStatus: %s", str)
}

type AudioFormat int

const (
	MP3 = iota
	MP4
	WAV
	FLAC
	AMR
	OGG
	WebM
)

var audioFormatStrings = []string{"MP3", "MP4", "WAV", "FLAC", "AMR", "OGG", "WebM"}

func (a AudioFormat) String() string {
	return audioFormatStrings[a]
}

func AudioFormatFromString(str string) (AudioFormat, error) {
	for i, s := range audioFormatStrings {
		if strings.EqualFold(s, str) { // Case insensitive comparison
			return AudioFormat(i), nil
		}
	}
	return 0, fmt.Errorf("invalid AudioFormat: %s", str)
}

// Transcript represents a session transcript in the system.
type Transcript struct {
	JobID              string
	AudioLocation      string
	AudioFormat        AudioFormat
	TranscriptLocation string
	SummaryLocation    string
	Status             TranscriptStatus
}
