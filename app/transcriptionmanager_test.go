package app

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/EdgarH78/dragonspeak-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	uuidString = "testUUID"
	testBucket = "testBucket"
)

type MockUUIDProvier struct {
}

func (m *MockUUIDProvier) NewUUID() string {
	return uuidString
}

type CapturedStartTranscriptionJobArgs struct {
	jobName        string
	audioLocation  string
	resultLocation string
	audioFormat    models.AudioFormat
}

type MockTranscriptionProvider struct {
	mock.Mock
	capturedArgs *CapturedStartTranscriptionJobArgs
}

func (m *MockTranscriptionProvider) StartTranscriptionJob(jobName, audioLocation, resultLocation string, audioFormat models.AudioFormat) error {
	args := m.Called(jobName, audioLocation, resultLocation, audioFormat)
	m.capturedArgs = &CapturedStartTranscriptionJobArgs{
		jobName:        jobName,
		audioLocation:  audioLocation,
		resultLocation: resultLocation,
		audioFormat:    audioFormat,
	}
	return args.Error(0)
}

type MockFileStore struct {
	files map[string][]byte
}

func NewMockFileStore() *MockFileStore {
	return &MockFileStore{
		files: map[string][]byte{},
	}
}

func (m *MockFileStore) UploadData(bucket, fileKey string, body io.Reader) error {
	key := fmt.Sprintf("%s/%s", bucket, fileKey)
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	m.files[key] = b
	return nil
}

func (m *MockFileStore) DownloadData(bucket, fileKey string, w io.WriterAt) (int64, error) {
	key := fmt.Sprintf("%s/%s", bucket, fileKey)
	b, ok := m.files[key]
	if !ok {
		return 0, models.EntityNotFound
	}
	w.WriteAt(b, 0)
	return int64(len(b)), nil
}

func (m *MockFileStore) GetContentFromPath(bucket, fileKey string) (string, bool) {
	key := fmt.Sprintf("%s/%s", bucket, fileKey)
	b, ok := m.files[key]
	if !ok {
		return "", false
	}
	return string(b), true
}

type MockTranscriptDb struct {
	mock.Mock
}

func (m *MockTranscriptDb) AddTranscriptToSessions(sessionID string, transcript models.Transcript) (*models.Transcript, error) {
	args := m.Called(sessionID, transcript)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), nil
}

func (m *MockTranscriptDb) GetTranscriptsForSessions(sessionID string) ([]models.Transcript, error) {
	args := m.Called(sessionID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Transcript), nil
}

func (m *MockTranscriptDb) GetTranscript(jobID string) (*models.Transcript, error) {
	args := m.Called(jobID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), nil
}

func TestSubmitTranscriptionJob(t *testing.T) {
	dbError := errors.New("db error")
	transcriptionJobError := errors.New("transcription job error")
	cases := []struct {
		description        string
		userID             string
		campaignID         string
		sessionID          string
		audioFormat        models.AudioFormat
		fileContent        string
		audioPath          string
		dbError            error
		dbResult           *models.Transcript
		fileStoreError     error
		transcriptionError error
		expectedDbRecord   *models.Transcript
		expectedError      error
		expectedResult     *models.Transcript
	}{
		{
			description: "transcription job is created",
			userID:      "user1",
			campaignID:  "campaign1",
			sessionID:   "session0",
			audioFormat: models.MP3,
			fileContent: "testaudio",
			audioPath:   "user1/campaign1/session0/audio-testUUID",
			expectedDbRecord: &models.Transcript{
				JobID:              "session0-testUUID",
				AudioLocation:      "user1/campaign1/session0/audio-testUUID",
				AudioFormat:        models.MP3,
				TranscriptLocation: "user1/campaign1/session0/transcript-testUUID",
				Status:             models.Transcribing,
			},
			dbResult: &models.Transcript{
				JobID:              "session0-testUUID",
				AudioLocation:      "user1/campaign1/session0/audio-testUUID",
				AudioFormat:        models.MP3,
				TranscriptLocation: "user1/campaign1/session0/transcript-testUUID",
				Status:             models.Transcribing,
			},
			expectedResult: &models.Transcript{
				JobID:              "session0-testUUID",
				AudioLocation:      "user1/campaign1/session0/audio-testUUID",
				AudioFormat:        models.MP3,
				TranscriptLocation: "user1/campaign1/session0/transcript-testUUID",
				Status:             models.Transcribing,
			},
		},
		{
			description:   "database returns an error, error returned",
			userID:        "user1",
			campaignID:    "campaign1",
			sessionID:     "session0",
			audioFormat:   models.MP3,
			fileContent:   "testaudio",
			audioPath:     "user1/campaign1/session0/audio-testUUID",
			dbError:       dbError,
			expectedError: dbError,
			expectedDbRecord: &models.Transcript{
				JobID:              "session0-testUUID",
				AudioLocation:      "user1/campaign1/session0/audio-testUUID",
				AudioFormat:        models.MP3,
				TranscriptLocation: "user1/campaign1/session0/transcript-testUUID",
				Status:             models.Transcribing,
			},
		},
		{
			description:        "transcription service returns an error, error returned",
			userID:             "user1",
			campaignID:         "campaign1",
			sessionID:          "session0",
			audioFormat:        models.MP3,
			fileContent:        "testaudio",
			audioPath:          "user1/campaign1/session0/audio-testUUID",
			transcriptionError: transcriptionJobError,
			expectedError:      transcriptionJobError,
			dbResult: &models.Transcript{
				JobID:              "session0-testUUID",
				AudioLocation:      "user1/campaign1/session0/audio-testUUID",
				AudioFormat:        models.MP3,
				TranscriptLocation: "user1/campaign1/session0/transcript-testUUID",
				Status:             models.Transcribing,
			},
			expectedDbRecord: &models.Transcript{
				JobID:              "session0-testUUID",
				AudioLocation:      "user1/campaign1/session0/audio-testUUID",
				AudioFormat:        models.MP3,
				TranscriptLocation: "user1/campaign1/session0/transcript-testUUID",
				Status:             models.Transcribing,
			},
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockTranscriptDb{}
			if c.dbError != nil {
				mockDb.On("AddTranscriptToSessions", mock.Anything, mock.Anything).Return(nil, c.dbError)
			} else {
				mockDb.On("AddTranscriptToSessions", c.sessionID, *c.expectedDbRecord).Return(c.dbResult, nil)
			}
			mockTranscriptionProvider := &MockTranscriptionProvider{}
			if c.transcriptionError != nil {
				mockTranscriptionProvider.On("StartTranscriptionJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(c.transcriptionError)
			} else {
				mockTranscriptionProvider.On("StartTranscriptionJob", c.expectedDbRecord.JobID, c.expectedDbRecord.AudioLocation, c.expectedDbRecord.TranscriptLocation, c.audioFormat).Return(nil)
			}
			mockFileStore := NewMockFileStore()
			mockUUIDProver := &MockUUIDProvier{}

			testManager := NewTranscriptionManager(testBucket, mockTranscriptionProvider, mockFileStore, mockDb, mockUUIDProver)

			result, err := testManager.SubmitTranscriptionJob(c.userID, c.campaignID, c.sessionID, c.audioFormat, strings.NewReader(c.fileContent))
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				assert.Equal(t, c.expectedResult.JobID, result.JobID)
				assert.Equal(t, c.expectedResult.AudioLocation, result.AudioLocation)
				assert.Equal(t, c.expectedResult.AudioFormat, result.AudioFormat)
				assert.Equal(t, c.expectedResult.Status, result.Status)
				assert.Equal(t, c.expectedResult.SummaryLocation, result.SummaryLocation)
				assert.Equal(t, c.expectedResult.TranscriptLocation, result.TranscriptLocation)

				uploadedContent, ok := mockFileStore.GetContentFromPath(testBucket, c.expectedDbRecord.AudioLocation)
				if !ok {
					t.Errorf("file content was not uploaded to the filestore")
				} else if uploadedContent != c.fileContent {
					t.Errorf("expected file content to be %s got %s", c.fileContent, uploadedContent)
				}
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
