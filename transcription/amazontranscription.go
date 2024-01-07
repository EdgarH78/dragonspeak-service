package transcription

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
)

// TranscriptionJobStatus is a custom type for transcription job statuses
type TranscriptionJobStatus int

// Constants for the transcription job statuses
const (
	TranscriptionJobStatusQueued TranscriptionJobStatus = iota
	TranscriptionJobStatusInProgress
	TranscriptionJobStatusFailed
	TranscriptionJobStatusCompleted
)

// statusStrings maps TranscriptionJobStatusType constants to their string representations
var statusStrings = []string{"QUEUED", "IN_PROGRESS", "FAILED", "COMPLETED"}

// String returns the string representation of a TranscriptionJobStatusType
func (s TranscriptionJobStatus) String() string {
	return statusStrings[s]
}

// FromString converts a string to a TranscriptionJobStatusType
func statusFromString(str string) (TranscriptionJobStatus, error) {
	for i, s := range statusStrings {
		if strings.EqualFold(s, str) { // Case insensitive comparison
			return TranscriptionJobStatus(i), nil
		}
	}
	return 0, fmt.Errorf("invalid TranscriptionJobStatusType: %s", str)
}

type AmazonTranscription struct {
	svc *transcribeservice.TranscribeService
}

func NewAmazonTranscription(sess *session.Session) *AmazonTranscription {
	return &AmazonTranscription{
		svc: transcribeservice.New(sess),
	}
}

func (t *AmazonTranscription) StartTranscriptionJob(jobName, audioLocation, resultLocation, mediaFormat string) error {
	// Start transcription job
	_, err := t.svc.StartTranscriptionJob(&transcribeservice.StartTranscriptionJobInput{
		TranscriptionJobName: &jobName,
		LanguageCode:         aws.String("en-US"),     // Set to the language of your audio file
		MediaFormat:          aws.String(mediaFormat), // Set to the format of your audio file
		Media: &transcribeservice.Media{
			MediaFileUri: aws.String(audioLocation),
		},
		OutputBucketName: aws.String(resultLocation),
	})

	return err
}

func (t *AmazonTranscription) GetTranscriptionJobStatus(jobName string) (TranscriptionJobStatus, error) {
	// Query the transcription job
	result, err := t.svc.GetTranscriptionJob(&transcribeservice.GetTranscriptionJobInput{
		TranscriptionJobName: &jobName,
	})
	if err != nil {
		return TranscriptionJobStatusFailed, err
	}

	status, err := statusFromString(*result.TranscriptionJob.TranscriptionJobStatus)
	if err != nil {
		return TranscriptionJobStatusFailed, err
	}

	return status, nil

}
