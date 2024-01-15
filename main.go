package main

import (
	"os"

	"github.com/EdgarH78/dragonspeak-service/app"
	"github.com/EdgarH78/dragonspeak-service/database"
	"github.com/EdgarH78/dragonspeak-service/filestorage"
	"github.com/EdgarH78/dragonspeak-service/presentation"
	"github.com/EdgarH78/dragonspeak-service/transcription"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
)

var (
	awsRegion  = os.Getenv("AWS_REGION")
	s3Bucket   = os.Getenv("S3_BUCKET")
	openAiKey  = os.Getenv("OPEN_AI_KEY")
	dbUser     = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbHost     = os.Getenv("DB_HOST")
	dbPort     = os.Getenv("DB_PORT")
	dbName     = os.Getenv("DB_NAME")
)

func main() {
	sqlConfig := database.SQLConfig{
		User:         dbUser,
		Password:     dbPassword,
		Host:         dbHost,
		Port:         dbPort,
		DatabaseName: dbName,
	}
	postgresDao, err := database.NewPostgresDao(sqlConfig)
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion), // Set your preferred region
	})
	s3Filestore := filestorage.NewS3Filestore(sess)
	amzTranscription := transcription.NewAmazonTranscription(sess, s3Bucket)
	campaignManager := app.NewCampaignManager(postgresDao)
	sessionManager := app.NewSessionManager(postgresDao)
	transciptionManager := app.NewTranscriptionManager(s3Bucket, amzTranscription, s3Filestore, postgresDao, &app.DefaultUUIDProvider{})
	userManager := app.NewUserManager(postgresDao)

	engine := gin.Default()
	api := presentation.NewHttpAPI(engine, userManager, campaignManager, sessionManager, transciptionManager)
	api.Run()
}
