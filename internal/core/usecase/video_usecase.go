package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain"
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

	// Step 1: Download and validate video
	localVideoPath, videoHash, err := uc.downloadAndValidateVideo(ctx, input.VideoKey)
	if err != nil {
		return uc.createErrorResponse("Failed to download or validate video", err)
	}
	defer func() {
		if localVideoPath != "" {
			uc.cleanupFile(ctx, localVideoPath, "temp video file")
		}
	}()

	// Step 2: Configure processing parameters
	cfg := uc.configureProcessing(input.Configuration, log)

	// Fail-fast: unsupported output_format
	if cfg.OutputFormat != "jpg" && cfg.OutputFormat != "png" {
		invErr := domain.NewInvalidInputError(fmt.Sprintf("unsupported output_format: %q (allowed: jpg, png)", cfg.OutputFormat))
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   invErr.Error(),
		}, invErr
	}

	// Step 3: Extract frames from video
	frameCount, zipPath, err := uc.extractFrames(ctx, localVideoPath, cfg)
	if err != nil {
		return uc.createErrorResponse("Failed to extract frames", err)
	}
	defer uc.cleanupFile(ctx, zipPath, "temp zip file")

	if frameCount == 0 {
		log.Warn("No frames extracted from video")
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   "No frames extracted from video",
		}, domain.NewInvalidInputError("no frames extracted from video")
	}

	// Step 4: Upload result using hash as filename
	outputKey, err := uc.uploadResult(ctx, zipPath, videoHash)
	if err != nil {
		return uc.createErrorResponse("Failed to upload result", err)
	}

	// Step 5: Cleanup - delete original video
	uc.deleteOriginalVideo(ctx, input.VideoKey)

	// Step 6: Update video status
	uc.updateVideoStatus(ctx, input.VideoId, input.UserId, videoHash, "PROCESSED")

	log.Info("Video processing completed successfully", "frame_count", frameCount, "output_key", outputKey, "hash", videoHash)
	return &dto.ProcessVideoOutput{
		Success:    true,
		Message:    fmt.Sprintf("Video processed successfully. %d frames extracted.", frameCount),
		OutputKey:  outputKey,
		FrameCount: frameCount,
		Hash:       videoHash,
	}, nil
}

// downloadAndValidateVideo downloads video, generates hash and validates it
func (uc *videoUseCase) downloadAndValidateVideo(ctx context.Context, videoKey string) (string, string, error) {
	log := uc.logger.WithContext(ctx).With("video_key", videoKey)
	log.Info("Starting video download")

	// Create temp file for video
	tempFile, err := uc.fileManager.CreateTempFile(ctx, "video_", ".mp4")
	if err != nil {
		log.Error("Failed to create temp file", "error", err)
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	log.Debug("Temp file created", "temp_file", tempFile)

	// Download video from storage
	reader, err := uc.videoGateway.Download(ctx, videoKey)
	if err != nil {
		log.Error("Failed to download from storage", "error", err)
		// Cleanup temp file on download error
		uc.cleanupFile(ctx, tempFile, "temp video file")
		var nErr *domain.NotFoundError
		if errors.As(err, &nErr) {
			return "", "", domain.NewNotFoundError(domain.ErrNotFound)
		}
		return "", "", fmt.Errorf("failed to download from storage: %w", err)
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil {
			log.Warn("Failed to close reader", "error", cerr)
		}
	}()
	log.Debug("Video reader obtained from storage")

	// Write to file and generate hash simultaneously
	hash, err := uc.writeFileAndGenerateHash(ctx, tempFile, reader)
	if err != nil {
		log.Error("Failed to write file and generate hash", "error", err)
		// Cleanup temp file on write error
		uc.cleanupFile(ctx, tempFile, "temp video file")
		return "", "", fmt.Errorf("failed to write file and generate hash: %w", err)
	}
	log.Info("Video downloaded successfully", "local_path", tempFile, "hash", hash)

	// Validate video format
	log.Info("Validating video format")
	if err := uc.videoProcessor.ValidateVideo(ctx, tempFile); err != nil {
		log.Error("Video validation failed", "error", err)
		// Cleanup temp file on validation error
		uc.cleanupFile(ctx, tempFile, "temp video file")
		return "", "", domain.NewValidationError(err)
	}
	log.Info("Video format validated successfully")

	return tempFile, hash, nil
}

// uploadResult uploads the processed video using hash as filename
func (uc *videoUseCase) uploadResult(ctx context.Context, zipPath, videoHash string) (string, error) {
	log := uc.logger.WithContext(ctx).With("zip_path", zipPath, "video_hash", videoHash)
	log.Info("Starting upload of processed video")

	reader, err := uc.fileManager.ReadFile(ctx, zipPath)
	if err != nil {
		log.Error("Failed to read zip file", "error", err)
		return "", fmt.Errorf("failed to read zip file: %w", err)
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil {
			log.Warn("Failed to close reader", "error", cerr)
		}
	}()

	size, err := uc.fileManager.GetFileSize(ctx, zipPath)
	if err != nil {
		log.Error("Failed to get file size", "error", err)
		return "", fmt.Errorf("failed to get file size: %w", err)
	}
	log.Debug("File size obtained", "size", size)

	// Generate output key using hash to avoid duplicates
	outputKey := uc.generateOutputKeyFromHash(videoHash)
	log.Debug("Generated output key from hash", "output_key", outputKey)

	_, err = uc.videoGateway.Upload(ctx, outputKey, reader, "application/zip", size)
	if err != nil {
		log.Error("Failed to upload to storage", "error", err)
		return "", fmt.Errorf("failed to upload to storage: %w", err)
	}
	log.Info("Upload completed successfully", "output_key", outputKey)

	return outputKey, nil
}

// generateOutputKeyFromHash creates output key using video hash to avoid duplicates
func (uc *videoUseCase) generateOutputKeyFromHash(videoHash string) string {
	return fmt.Sprintf("processed/%s.zip", videoHash)
}

// writeFileAndGenerateHash writes content to file while generating SHA-256 hash
func (uc *videoUseCase) writeFileAndGenerateHash(ctx context.Context, filePath string, reader io.Reader) (string, error) {
	// Create a tee reader to calculate hash while writing
	hasher := sha256.New()
	teeReader := io.TeeReader(reader, hasher)

	// Write to file
	if err := uc.fileManager.WriteToFile(ctx, filePath, teeReader); err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	// Generate hash
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash, nil
}

// configureProcessing sets up processing configuration
func (uc *videoUseCase) configureProcessing(inputConfig *dto.ProcessingConfigInput, log logger.Logger) entity.ProcessingConfig {
	var cfg entity.ProcessingConfig
	if inputConfig == nil {
		cfg = entity.ProcessingConfig{FrameRate: 1.0, OutputFormat: "jpg"}
		log.Info("Using default configuration")
	} else {
		cfg = entity.ProcessingConfig{FrameRate: inputConfig.FrameRate, OutputFormat: inputConfig.OutputFormat}
		log.Info("Using custom configuration", "frame_rate", cfg.FrameRate, "output_format", cfg.OutputFormat)
	}

	if cfg.FrameRate <= 0 {
		cfg.FrameRate = 1.0
	}
	cfg.OutputFormat = strings.ToLower(strings.TrimSpace(cfg.OutputFormat))
	if cfg.OutputFormat == "jpeg" {
		cfg.OutputFormat = "jpg"
	}
	return cfg
}

// extractFrames processes video and extracts frames
func (uc *videoUseCase) extractFrames(ctx context.Context, videoPath string, cfg entity.ProcessingConfig) (int, string, error) {
	log := uc.logger.WithContext(ctx)
	log.Info("Starting frame extraction")

	frameCount, zipPath, err := uc.videoProcessor.ProcessVideo(ctx, videoPath, cfg.FrameRate, cfg.OutputFormat)
	if err != nil {
		log.Error("Failed to process video", "error", err)
		return 0, "", fmt.Errorf("failed to process video: %w", err)
	}

	log.Info("Frame extraction completed", "frame_count", frameCount, "zip_path", zipPath)
	return frameCount, zipPath, nil
}

// createErrorResponse creates standardized error response
func (uc *videoUseCase) createErrorResponse(message string, err error) (*dto.ProcessVideoOutput, error) {
	var nErr *domain.NotFoundError
	var vErr *domain.ValidationError

	if errors.As(err, &nErr) {
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   fmt.Sprintf("%s: %v", message, err),
		}, nErr
	}

	if errors.As(err, &vErr) {
		return &dto.ProcessVideoOutput{
			Success: false,
			Message: "Processing failed",
			Error:   fmt.Sprintf("%s: %v", message, err),
		}, vErr
	}

	return &dto.ProcessVideoOutput{
		Success: false,
		Message: "Processing failed",
		Error:   fmt.Sprintf("%s: %v", message, err),
	}, domain.NewInternalError(err)
}

// cleanupFile safely deletes temporary files
func (uc *videoUseCase) cleanupFile(ctx context.Context, filePath, fileType string) {
	if err := uc.fileManager.DeleteFile(ctx, filePath); err != nil {
		uc.logger.WithContext(ctx).Warn("Failed to delete "+fileType, "path", filePath, "error", err)
	}
}

// deleteOriginalVideo removes the original video from storage
func (uc *videoUseCase) deleteOriginalVideo(ctx context.Context, videoKey string) {
	log := uc.logger.WithContext(ctx)
	log.Info("Deleting original video from storage")
	if err := uc.videoGateway.Delete(ctx, videoKey); err != nil {
		log.Warn("Failed to delete original video", "error", err, "video_key", videoKey)
	} else {
		log.Info("Original video deleted successfully")
	}
}

func (uc *videoUseCase) updateVideoStatus(ctx context.Context, videoId string, userId string, videoHash, status string) {
	log := uc.logger.WithContext(ctx)
	log.Info("Updating video status", "video_key", videoId, "status", status)
	if err := uc.videoGateway.UpdateStatus(ctx, videoId, userId, videoHash, status); err != nil {
		log.Warn("Failed to update video status", "error", err, "video_key", videoId, "status", status)
	} else {
		log.Info("Video status updated successfully", "video_key", videoId, "status", status)
	}
}
