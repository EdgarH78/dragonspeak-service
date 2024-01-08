package app

import (
	"fmt"
	"io"

	"github.com/EdgarH78/dragonspeak-service/models"
)

type transcriptionProvider interface {
	StartTranscriptionJob(jobName, audioLocation, resultLocation string, audioFormat models.AudioFormat) error
}

type fileStore interface {
	UploadData(bucket, fileKey string, body io.Reader) error
	DownloadData(bucket, fileKey string, w io.WriterAt) (int64, error)
}

type transcriptionDb interface {
	AddTranscriptToSessions(sessionID string, transcript models.Transcript) (*models.Transcript, error)
	GetTranscriptsForSessions(sessionID string) ([]models.Transcript, error)
	GetTranscript(jobID string) (*models.Transcript, error)
}

type uuidProvider interface {
	NewUUID() string
}

type TranscriptionManager struct {
	bucket                string
	transcriptionProvider transcriptionProvider
	fileStore             fileStore
	transcriptionDb       transcriptionDb
	uuidProvider          uuidProvider
}

func NewTranscriptionManager(bucket string, transcriptionProvider transcriptionProvider, fileSfileStore fileStore, tratranscriptionDb transcriptionDb, uuidProvider uuidProvider) *TranscriptionManager {
	return &TranscriptionManager{
		bucket:                bucket,
		transcriptionProvider: transcriptionProvider,
		fileStore:             fileSfileStore,
		transcriptionDb:       tratranscriptionDb,
		uuidProvider:          uuidProvider,
	}
}

func (t *TranscriptionManager) SubmitTranscriptionJob(userID, campaignID, sessionID string, audioFormat models.AudioFormat, audioFile io.Reader) (*models.Transcript, error) {
	jobID := fmt.Sprintf("%s-%s", sessionID, t.uuidProvider.NewUUID())
	audioLocation := fmt.Sprintf("%s/%s/%s/audio-%s", userID, campaignID, sessionID, t.uuidProvider.NewUUID())
	transcriptLocation := fmt.Sprintf("%s/%s/%s/transcript-%s", userID, campaignID, sessionID, t.uuidProvider.NewUUID())
	transcriptionJob := models.Transcript{
		JobID:              jobID,
		AudioLocation:      audioLocation,
		AudioFormat:        audioFormat,
		TranscriptLocation: transcriptLocation,
		Status:             models.Transcribing,
	}

	_, err := t.transcriptionDb.AddTranscriptToSessions(sessionID, transcriptionJob)
	if err != nil {
		return nil, err
	}
	err = t.fileStore.UploadData(t.bucket, audioLocation, audioFile)
	if err != nil {
		return nil, err
	}
	err = t.transcriptionProvider.StartTranscriptionJob(jobID, audioLocation, transcriptLocation, audioFormat)
	if err != nil {
		return nil, err
	}

	return &transcriptionJob, nil
}

func (t *TranscriptionManager) GetTranscriptJob(jobID string) (*models.Transcript, error) {
	return nil, fmt.Errorf("not implemented")
}

func (t *TranscriptionManager) DownloadTranscript(jobID string, w io.WriterAt) error {
	return fmt.Errorf("not implemented")
}