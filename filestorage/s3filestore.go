package filestorage

import (
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Define a struct to hold the S3 uploader
type S3Filestore struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewS3Uploader creates a new S3 Uploader instance
func NewS3Filestore(sess *session.Session) *S3Filestore {
	return &S3Filestore{
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

func (f *S3Filestore) UploadData(bucket, fileKey string, body io.Reader) error {
	_, err := f.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath.Base(fileKey)),
		Body:   body,
	})

	return err
}

func (f *S3Filestore) DownloadData(bucket, fileKey string, w io.WriterAt) (int64, error) {

	// Download the item from the bucket. If an error occurs, log it and exit.
	numBytes, err := f.downloader.Download(w,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileKey),
		})
	if err != nil {
		return 0, err
	}

	return numBytes, nil
}
