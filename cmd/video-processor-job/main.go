package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/controller"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/gateway"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/presenter"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/usecase"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/datasource"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/service"
)

func main() {
	// Get configuration from environment variables (K8s job format)
	key := os.Getenv("K8S_JOB_ENV_VIDEO_KEY")                        // S3 key (path) of the video file
	outputFormat := getEnv("K8S_JOB_ENV_VIDEO_EXPORT_FORMAT", "jpg") // Output format (jpg or png)
	frameRateStr := getEnv("K8S_JOB_ENV_VIDEO_EXPORT_FPS", "1.0")    // Frame rate for extraction

	// Parse frame rate
	frameRate := 1.0
	if fr, err := strconv.ParseFloat(frameRateStr, 64); err == nil {
		frameRate = fr
	} else {
		fmt.Printf("Warning: Invalid FRAME_RATE '%s', using default 1.0\n", frameRateStr)
	}

	if key == "" {
		fmt.Println("Required environment variables:")
		fmt.Println("  K8S_JOB_ENV_VIDEO_KEY=<s3-key-path>        # S3 key of the video file")
		fmt.Println("  K8S_JOB_ENV_VIDEO_BUCKET=<bucket-name>     # S3 bucket for input videos")
		fmt.Println("  K8S_JOB_ENV_PROCESSED_BUCKET=<bucket-name> # S3 bucket for output files")
		fmt.Println("")
		fmt.Println("Optional environment variables:")
		fmt.Println("  K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg|png    # Output format (default: jpg)")
		fmt.Println("  K8S_JOB_ENV_VIDEO_EXPORT_FPS=1.0           # Frame extraction rate (default: 1.0)")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  export K8S_JOB_ENV_VIDEO_KEY=videos/sample.mp4")
		fmt.Println("  export K8S_JOB_ENV_VIDEO_EXPORT_FORMAT=jpg")
		fmt.Println("  export K8S_JOB_ENV_VIDEO_EXPORT_FPS=2.0")
		fmt.Println("  ./video-processor-job")
		os.Exit(1)
	}

	// Initialize context with trace id, then logger
	ctx := context.Background()
	traceID := generateTraceID()
	ctx = logger.SetTraceIDOnContext(ctx, traceID)

	logger := logger.NewSlogLogger().With("trace_id", traceID)
	logger.Info("Starting Video Processor standalone application")

	// Initialize AWS config
	logger.Info("Loading AWS configuration")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load AWS config", "error", err)
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Initialize S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Get bucket names from environment variables
	videoBucket := getEnv("K8S_JOB_ENV_VIDEO_BUCKET", "video-processor-raw-videos")
	processedBucket := getEnv("K8S_JOB_ENV_PROCESSED_BUCKET", "video-processor-processed-images")
	logger.Info("Environment configuration loaded", "video_bucket", videoBucket, "processed_bucket", processedBucket)

	// Initialize infrastructure layer
	logger.Info("Initializing infrastructure layer")
	storageDataSource := datasource.NewS3StorageDataSource(s3Client, videoBucket, processedBucket)
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
		"key", key,
		"format", outputFormat,
		"fps", frameRate)

	input := dto.ProcessVideoInput{
		VideoKey: key,
		Configuration: &dto.ProcessingConfigInput{
			FrameRate:    frameRate,
			OutputFormat: outputFormat,
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

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// generateTraceID creates a random 16-byte hex string for tracing
func generateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
