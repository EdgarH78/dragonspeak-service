package app

import (
	"context"
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

// BufferWriterAt is an in-memory buffer that implements io.WriterAt.
type BufferWriterAt struct {
	buf []byte
}

// NewBufferWriterAt creates a new BufferWriterAt with a given size.
func NewBufferWriterAt(size int) *BufferWriterAt {
	return &BufferWriterAt{buf: make([]byte, size)}
}

// WriteAt implements the io.WriterAt interface for BufferWriterAt.
func (b *BufferWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 || int(off) > len(b.buf) {
		return 0, io.ErrShortWrite
	}
	n = copy(b.buf[off:], p)
	if n < len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (b *BufferWriterAt) GetData() []byte {
	return b.buf
}

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

func (m *MockTranscriptDb) AddTranscriptToSession(ctx context.Context, sessionID string, transcript models.Transcript) (*models.Transcript, error) {
	args := m.Called(ctx, sessionID, transcript)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transcript), nil
}

func (m *MockTranscriptDb) GetTranscriptsForSession(ctx context.Context, sessionID string) ([]models.Transcript, error) {
	args := m.Called(ctx, sessionID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Transcript), nil
}

func (m *MockTranscriptDb) GetTranscript(ctx context.Context, jobID string) (*models.Transcript, error) {
	args := m.Called(ctx, jobID)
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
				mockDb.On("AddTranscriptToSession", mock.Anything, mock.Anything, mock.Anything).Return(nil, c.dbError)
			} else {
				mockDb.On("AddTranscriptToSession", mock.Anything, c.sessionID, *c.expectedDbRecord).Return(c.dbResult, nil)
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

			result, err := testManager.SubmitTranscriptionJob(context.Background(), c.userID, c.campaignID, c.sessionID, c.audioFormat, strings.NewReader(c.fileContent))
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

func TestGetTranscriptionJob(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		jobID          string
		dbError        error
		dbResult       *models.Transcript
		expectedError  error
		expectedResult *models.Transcript
	}{
		{
			description: "Transcript is retrieved",
			jobID:       "job-1",
			dbResult: &models.Transcript{
				JobID:              "job-1",
				AudioLocation:      "audio.wav",
				AudioFormat:        models.WAV,
				TranscriptLocation: "transcript.txt",
				SummaryLocation:    "summary.txt",
				Status:             models.Done,
			},
			expectedResult: &models.Transcript{
				JobID:              "job-1",
				AudioLocation:      "audio.wav",
				AudioFormat:        models.WAV,
				TranscriptLocation: "transcript.txt",
				SummaryLocation:    "summary.txt",
				Status:             models.Done,
			},
		},
		{
			description:   "database returns error",
			jobID:         "job-1",
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockTranscriptDb{}
			if c.dbError != nil {
				mockDb.On("GetTranscript", mock.Anything, c.jobID).Return(nil, c.dbError)
			} else {
				mockDb.On("GetTranscript", mock.Anything, c.jobID).Return(c.dbResult, nil)
			}
			mockFileStore := NewMockFileStore()
			mockUUIDProver := &MockUUIDProvier{}
			mockTranscriptionProvider := &MockTranscriptionProvider{}

			testManager := NewTranscriptionManager(testBucket, mockTranscriptionProvider, mockFileStore, mockDb, mockUUIDProver)

			result, err := testManager.GetTranscriptJob(context.Background(), c.jobID)
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

func TestGetTranscriptsForSession(t *testing.T) {
	dbError := errors.New("db error")
	cases := []struct {
		description    string
		sessionID      string
		dbError        error
		dbResult       []models.Transcript
		expectedError  error
		expectedResult []models.Transcript
	}{
		{
			description: "Transcript is retrieved",
			sessionID:   "session-1",
			dbResult: []models.Transcript{
				{
					JobID:              "job-1",
					AudioLocation:      "audio.wav",
					AudioFormat:        models.WAV,
					TranscriptLocation: "transcript.txt",
					SummaryLocation:    "summary.txt",
					Status:             models.Done,
				},
			},
			expectedResult: []models.Transcript{
				{
					JobID:              "job-1",
					AudioLocation:      "audio.wav",
					AudioFormat:        models.WAV,
					TranscriptLocation: "transcript.txt",
					SummaryLocation:    "summary.txt",
					Status:             models.Done,
				},
			},
		},
		{
			description:   "database returns error",
			sessionID:     "session-1",
			dbError:       dbError,
			expectedError: dbError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockTranscriptDb{}
			if c.dbError != nil {
				mockDb.On("GetTranscriptsForSession", mock.Anything, c.sessionID).Return(nil, c.dbError)
			} else {
				mockDb.On("GetTranscriptsForSession", mock.Anything, c.sessionID).Return(c.dbResult, nil)
			}
			mockFileStore := NewMockFileStore()
			mockUUIDProver := &MockUUIDProvier{}
			mockTranscriptionProvider := &MockTranscriptionProvider{}

			testManager := NewTranscriptionManager(testBucket, mockTranscriptionProvider, mockFileStore, mockDb, mockUUIDProver)

			result, err := testManager.GetTranscriptsForSession(context.Background(), c.sessionID)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			}

			if c.expectedResult != nil {
				if len(c.expectedResult) != len(result) {
					t.Errorf("expected %d results got %d", len(c.expectedResult), len(result))
				}
				for i, expected := range c.expectedResult {
					actual := result[i]
					assert.Equal(t, expected.JobID, actual.JobID)
					assert.Equal(t, expected.AudioLocation, actual.AudioLocation)
					assert.Equal(t, expected.AudioFormat, actual.AudioFormat)
					assert.Equal(t, expected.Status, actual.Status)
					assert.Equal(t, expected.SummaryLocation, actual.SummaryLocation)
					assert.Equal(t, expected.TranscriptLocation, actual.TranscriptLocation)
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

func TestDownloadTranscript(t *testing.T) {
	dbError := fmt.Errorf("db error")
	fileStoreError := fmt.Errorf("file store error")

	cases := []struct {
		description     string
		jobID           string
		filecontent     string
		transcript      *models.Transcript
		dbError         error
		fileStoreError  error
		expectedError   error
		expectedContent string
	}{
		{
			description:     "content downloaded",
			jobID:           "jobId-1",
			filecontent:     "this is a test",
			expectedContent: "this is a test",
			transcript: &models.Transcript{
				JobID:              "jobId-1",
				AudioLocation:      "audio.wav",
				AudioFormat:        models.WAV,
				TranscriptLocation: "transcript.txt",
				SummaryLocation:    "",
				Status:             models.Done,
			},
		},
		{
			description:     "transcript status is Transcribing, Conflicted error returned",
			jobID:           "jobId-1",
			filecontent:     "this is a test",
			expectedContent: "this is a test",
			transcript: &models.Transcript{
				JobID:              "jobId-1",
				AudioLocation:      "audio.wav",
				AudioFormat:        models.WAV,
				TranscriptLocation: "transcript.txt",
				SummaryLocation:    "",
				Status:             models.Transcribing,
			},
			expectedError: models.Conflicted,
		},
		{
			description:     "database returns an error, error returned",
			jobID:           "jobId-1",
			filecontent:     "this is a test",
			expectedContent: "this is a test",
			transcript: &models.Transcript{
				JobID:              "jobId-1",
				AudioLocation:      "audio.wav",
				AudioFormat:        models.WAV,
				TranscriptLocation: "transcript.txt",
				SummaryLocation:    "",
				Status:             models.Done,
			},
			dbError:       dbError,
			expectedError: dbError,
		},
		{
			description:     "filestore returns an error, error returned",
			jobID:           "jobId-1",
			filecontent:     "this is a test",
			expectedContent: "this is a test",
			transcript: &models.Transcript{
				JobID:              "jobId-1",
				AudioLocation:      "audio.wav",
				AudioFormat:        models.WAV,
				TranscriptLocation: "transcript.txt",
				SummaryLocation:    "",
				Status:             models.Done,
			},
			fileStoreError: fileStoreError,
			expectedError:  fileStoreError,
		},
	}

	// Iterate through test cases
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			mockDb := &MockTranscriptDb{}
			if c.dbError != nil {
				mockDb.On("GetTranscript", mock.Anything, c.jobID).Return(nil, c.dbError)
			} else {
				mockDb.On("GetTranscript", mock.Anything, c.jobID).Return(c.transcript, nil)
			}
			mockFileStore := NewMockFileStore()
			mockFileStore.UploadData(testBucket, "transcript.txt", strings.NewReader(c.filecontent))
			mockUUIDProver := &MockUUIDProvier{}
			mockTranscriptionProvider := &MockTranscriptionProvider{}

			testManager := NewTranscriptionManager(testBucket, mockTranscriptionProvider, mockFileStore, mockDb, mockUUIDProver)

			bufferWriter := NewBufferWriterAt(len([]byte(c.filecontent)))
			err := testManager.DownloadTranscript(context.Background(), c.jobID, bufferWriter)
			if err != nil && c.expectedError == nil {
				t.Errorf("unexpected error returned: %s", err)
				return
			} else if err != nil {
				if !errors.Is(err, c.expectedError) {
					t.Errorf("expected error to be %s got %s", c.expectedError, err)
				}
			} else {
				result := string(bufferWriter.GetData())
				if result != c.expectedContent {
					t.Errorf("expected file content to be %s got %s", c.filecontent, result)
				}
			}
		})
	}

}
