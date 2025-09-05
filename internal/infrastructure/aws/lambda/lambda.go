package lambda

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain/entity"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
)

// Handler handles Lambda function invocations
type Handler struct {
	videoController port.VideoController
	logger          logger.Logger
}

// NewHandler creates a new Lambda handler
func NewHandler(videoController port.VideoController, logger logger.Logger) *Handler {
	return &Handler{
		videoController: videoController,
		logger:          logger,
	}
}

// Handle processes Lambda events
func (h *Handler) Handle(ctx context.Context, event json.RawMessage) (interface{}, error) {
	log := h.logger.WithContext(ctx)
	log.Info("Lambda function invoked")

	// Parse direct invocation event
	var lambdaEvent LambdaEvent
	if err := json.Unmarshal(event, &lambdaEvent); err != nil {
		log.Error("Failed to parse Lambda event", "error", err)
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	log.Debug("Lambda event parsed successfully", "video_key", lambdaEvent.VideoKey)

	if lambdaEvent.VideoKey == "" {
		log.Error("Missing video_key in Lambda event")
		return nil, fmt.Errorf("video_key is required")
	}

	// Convert to DTO
	input := dto.ProcessVideoInput{
		VideoKey:      lambdaEvent.VideoKey,
		Configuration: lambdaEvent.Configuration,
	}

	// Process video
	log.Info("Delegating to video controller", "video_key", input.VideoKey)
	result, err := h.videoController.ProcessVideo(ctx, input)
	if err != nil {
		log.Error("Video processing failed in controller", "error", err)
		// Return error response with proper status code
		return LambdaResponse{
			StatusCode: 500,
			Body:       string(result),
		}, nil
	}

	// Return success response
	log.Info("Lambda function completed successfully")
	return LambdaResponse{
		StatusCode: 200,
		Body:       string(result),
	}, nil
}

// LambdaEvent represents the input event for Lambda
type LambdaEvent struct {
	VideoKey      string                   `json:"video_key"`
	Configuration *entity.ProcessingConfig `json:"configuration,omitempty"`
}

// LambdaResponse represents the Lambda function response
type LambdaResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}
