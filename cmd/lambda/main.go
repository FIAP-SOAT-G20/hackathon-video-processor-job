package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/controller"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/gateway"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/presenter"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/usecase"
	lambdaHandler "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/aws/lambda"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/datasource"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/service"
)

func main() {
	// Initialize logger
	log := logger.NewSlogLogger()
	log.Info("Starting Video Processor Lambda function")

	// Initialize AWS config
	ctx := context.Background()
	log.Info("Loading AWS configuration")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Failed to load AWS config", "error", err)
		panic(err)
	}

	// Initialize S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Get bucket names from environment variables
	videoBucket := getEnv("VIDEO_BUCKET", "video-processor-raw-videos")
	processedBucket := getEnv("PROCESSED_BUCKET", "video-processor-processed-images")
	log.Info("Environment configuration loaded", "video_bucket", videoBucket, "processed_bucket", processedBucket)

	// Initialize infrastructure layer
	log.Info("Initializing infrastructure layer")
	storageDataSource := datasource.NewS3StorageDataSource(s3Client, videoBucket, processedBucket)
	fileManager := service.NewLocalFileService()
	videoProcessor := service.NewFFmpegService(fileManager)

	// Initialize adapter layer
	log.Info("Initializing adapter layer")
	videoGateway := gateway.NewVideoGateway(storageDataSource)
	videoPresenter := presenter.NewVideoJsonPresenter()

	// Initialize core layer
	log.Info("Initializing core layer")
	videoUseCase := usecase.NewVideoUseCase(videoGateway, videoProcessor, fileManager, log)
	videoController := controller.NewVideoController(videoUseCase, videoPresenter, log)

	// Initialize Lambda handler
	log.Info("Initializing Lambda handler")
	handler := lambdaHandler.NewHandler(videoController, log)

	// Start Lambda function
	log.Info("Lambda function initialized successfully, ready to handle requests")
	lambda.Start(handler.Handle)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
