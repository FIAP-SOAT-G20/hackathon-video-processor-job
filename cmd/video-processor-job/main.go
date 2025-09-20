package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/controller"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/gateway"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/presenter"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/usecase"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/config"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/datasource"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate required fields
	if err := cfg.ValidateRequiredFields(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize context with trace id, then logger
	ctx := context.Background()
	traceID := generateTraceID()
	ctx = logger.SetTraceIDOnContext(ctx, traceID)

	logger := logger.NewSlogLogger().With("trace_id", traceID)
	logger.Info("Starting Video Processor standalone application")

	// Initialize AWS config with explicit credentials
	logger.Info("Loading AWS configuration", "region", cfg.AWS.Region)
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWS.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKey,
			cfg.AWS.SecretAccessKey,
			cfg.AWS.SessionToken,
		)),
	)
	if err != nil {
		logger.Error("Failed to load AWS config", "error", err)
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Initialize S3 client
	s3Client := s3.NewFromConfig(awsCfg)

	logger.Info("Environment configuration loaded",
		"video_bucket", cfg.Video.Bucket,
		"processed_bucket", cfg.Video.ProcessedBucket)

	// Initialize infrastructure layer
	logger.Info("Initializing infrastructure layer")
	storageDataSource := datasource.NewS3StorageDataSource(s3Client, cfg.Video.Bucket, cfg.Video.ProcessedBucket)
	fileManager := service.NewLocalFileService()
	videoProcessor := service.NewFFmpegService(fileManager)

	// Initialize adapter layer
	logger.Info("Initializing adapter layer")
	videoGateway := gateway.NewVideoGateway(storageDataSource)
	videoPresenter := presenter.NewVideoJsonPresenter()

	// Initialize core layer
	logger.Info("Initializing core layer")
	videoUseCase := usecase.NewVideoUseCase(videoGateway, videoProcessor, fileManager, logger)
	videoController := controller.NewVideoController(videoUseCase, videoPresenter, logger)

	// Create processing input using DTOs
	logger.Info("Processing video",
		"key", cfg.Video.Key,
		"format", cfg.Video.ExportFormat,
		"fps", cfg.Video.ExportFPS)

	input := dto.ProcessVideoInput{
		VideoKey: cfg.Video.Key,
		Configuration: &dto.ProcessingConfigInput{
			FrameRate:    cfg.Video.ExportFPS,
			OutputFormat: cfg.Video.ExportFormat,
		},
	}

	// Process the video
	result, err := videoController.ProcessVideo(ctx, input)
	if err != nil {
		logger.Error("Failed to process video", "error", err)
		log.Fatalf("Failed to process video: %v", err)
	}

	// Print result
	fmt.Println("Video processed successfully!")
	fmt.Printf("Result: %s\n", string(result))
}

// generateTraceID creates a random 16-byte hex string for tracing
func generateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
