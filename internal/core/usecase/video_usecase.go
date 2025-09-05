package usecase

import (
	"context"
	"fmt"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain/entity"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
)

type videoUseCase struct {
	videoGateway   port.VideoGateway
	videoProcessor port.VideoProcessor
	fileManager    port.FileManager
	logger         logger.Logger
}

func NewVideoUseCase(
	videoGateway port.VideoGateway,
	videoProcessor port.VideoProcessor,
	fileManager port.FileManager,
	logger logger.Logger,
) port.VideoUseCase {
	return &videoUseCase{
		videoGateway:   videoGateway,
		videoProcessor: videoProcessor,
		fileManager:    fileManager,
		logger:         logger,
	}
}

func (uc *videoUseCase) ProcessVideo(ctx context.Context, input dto.ProcessVideoInput) (*dto.ProcessVideoOutput, error) {
	log := uc.logger.WithContext(ctx).With("video_key", input.VideoKey)
	log.Info("Starting video processing")
	// Download video from storage
	log.Info("Starting video download")
	localVideoPath, err := uc.downloadVideoToLocal(ctx, input.VideoKey)
	if err != nil {
		log.Error("Failed to download video", "error", err)
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   fmt.Sprintf("Failed to download video: %v", err),
		}, err
	}
	defer uc.fileManager.DeleteFile(ctx, localVideoPath)
	log.Info("Video downloaded successfully", "local_path", localVideoPath)

	// Validate video format
	log.Info("Validating video format")
	if err := uc.videoProcessor.ValidateVideo(ctx, localVideoPath); err != nil {
		log.Error("Video validation failed", "error", err)
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   fmt.Sprintf("Invalid video format: %v", err),
		}, err
	}
	log.Info("Video format validated successfully")

	// Set default configuration if not provided
	config := input.Configuration
	if config == nil {
		config = &entity.ProcessingConfig{
			FrameRate:    1.0,
			OutputFormat: "png",
		}
		log.Info("Using default configuration")
	} else {
		log.Info("Using custom configuration", "frame_rate", config.FrameRate, "output_format", config.OutputFormat)
	}

	// Process video and extract frames
	log.Info("Starting frame extraction")
	_, frameCount, zipPath, err := uc.videoProcessor.ProcessVideo(ctx, localVideoPath, config.FrameRate)
	if err != nil {
		log.Error("Failed to process video", "error", err)
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   fmt.Sprintf("Failed to process video: %v", err),
		}, err
	}
	defer uc.fileManager.DeleteFile(ctx, zipPath)
	log.Info("Frame extraction completed", "frame_count", frameCount, "zip_path", zipPath)

	if frameCount == 0 {
		log.Warn("No frames extracted from video")
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   "No frames extracted from video",
		}, fmt.Errorf("no frames extracted from video")
	}

	// Upload ZIP to storage
	log.Info("Starting upload of processed video")
	outputKey, err := uc.uploadResultToStorage(ctx, zipPath, input.VideoKey)
	if err != nil {
		log.Error("Failed to upload result", "error", err)
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   fmt.Sprintf("Failed to upload result: %v", err),
		}, err
	}
	log.Info("Upload completed successfully", "output_key", outputKey)

	// Delete original video from storage
	log.Info("Deleting original video from storage")
	if err := uc.videoGateway.Delete(ctx, input.VideoKey); err != nil {
		log.Warn("Failed to delete original video", "error", err, "video_key", input.VideoKey)
	} else {
		log.Info("Original video deleted successfully")
	}

	log.Info("Video processing completed successfully", "frame_count", frameCount, "output_key", outputKey)
	return &dto.ProcessVideoOutput{
		Success:    true,
		Message:    fmt.Sprintf("Video processed successfully. %d frames extracted.", frameCount),
		OutputKey:  outputKey,
		FrameCount: frameCount,
	}, nil
}

func (uc *videoUseCase) downloadVideoToLocal(ctx context.Context, videoKey string) (string, error) {
	log := uc.logger.WithContext(ctx).With("video_key", videoKey)
	tempFile, err := uc.fileManager.CreateTempFile(ctx, "video_", ".mp4")
	if err != nil {
		log.Error("Failed to create temp file", "error", err)
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	log.Debug("Temp file created", "temp_file", tempFile)

	reader, err := uc.videoGateway.Download(ctx, videoKey)
	if err != nil {
		log.Error("Failed to download from storage", "error", err)
		return "", fmt.Errorf("failed to download from storage: %w", err)
	}
	defer reader.Close()
	log.Debug("Video reader obtained from storage")

	if err := uc.fileManager.WriteToFile(ctx, tempFile, reader); err != nil {
		log.Error("Failed to write to local file", "error", err)
		return "", fmt.Errorf("failed to write to local file: %w", err)
	}
	log.Debug("Video written to local file successfully")

	return tempFile, nil
}

func (uc *videoUseCase) uploadResultToStorage(ctx context.Context, zipPath, originalVideoKey string) (string, error) {
	log := uc.logger.WithContext(ctx).With("original_video_key", originalVideoKey, "zip_path", zipPath)
	reader, err := uc.fileManager.ReadFile(ctx, zipPath)
	if err != nil {
		log.Error("Failed to read zip file", "error", err)
		return "", fmt.Errorf("failed to read zip file: %w", err)
	}
	defer reader.Close()

	size, err := uc.fileManager.GetFileSize(ctx, zipPath)
	if err != nil {
		log.Error("Failed to get file size", "error", err)
		return "", fmt.Errorf("failed to get file size: %w", err)
	}
	log.Debug("File size obtained", "size", size)

	outputKey := uc.generateOutputKey(originalVideoKey)
	log.Debug("Generated output key", "output_key", outputKey)

	_, err = uc.videoGateway.Upload(ctx, outputKey, reader, "application/zip", size)
	if err != nil {
		log.Error("Failed to upload to storage", "error", err)
		return "", fmt.Errorf("failed to upload to storage: %w", err)
	}
	log.Debug("Upload to storage completed successfully")

	return outputKey, nil
}

func (uc *videoUseCase) generateOutputKey(originalVideoKey string) string {
	baseKey := originalVideoKey
	if lastDot := uc.lastIndexOf(baseKey, "."); lastDot != -1 {
		baseKey = baseKey[:lastDot]
	}

	return fmt.Sprintf("processed/%s_frames.zip", baseKey)
}

func (uc *videoUseCase) lastIndexOf(s, substr string) int {
	idx := -1
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			idx = i
		}
	}
	return idx
}
