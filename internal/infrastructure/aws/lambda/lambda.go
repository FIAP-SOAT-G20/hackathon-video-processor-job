package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain"
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
		body := toErrorBody("Bad request", fmt.Sprintf("failed to parse event: %v", err))
		return LambdaResponse{StatusCode: 400, Body: body}, nil
	}
	log.Debug("Lambda event parsed successfully", "video_key", lambdaEvent.VideoKey)

	if lambdaEvent.VideoKey == "" {
		log.Error("Missing video_key in Lambda event")
		body := toErrorBody("Bad request", "video_key is required")
		return LambdaResponse{StatusCode: 400, Body: body}, nil
	}

	// Convert to DTO (map payload to application DTO)
	var cfg *dto.ProcessingConfigInput
	if lambdaEvent.Configuration != nil {
		cfg = &dto.ProcessingConfigInput{
			FrameRate:    lambdaEvent.Configuration.FrameRate,
			OutputFormat: lambdaEvent.Configuration.OutputFormat,
		}
	}
	input := dto.ProcessVideoInput{
		VideoKey:      lambdaEvent.VideoKey,
		Configuration: cfg,
	}

	// Process video
	log.Info("Delegating to video controller", "video_key", input.VideoKey)
	result, err := h.videoController.ProcessVideo(ctx, input)
	if err != nil {
		log.Error("Video processing failed in controller", "error", err)
		// Map domain error types to HTTP status codes
		status := 500
		var vErr *domain.ValidationError
		var iErr *domain.InvalidInputError
		var nErr *domain.NotFoundError
		if errors.As(err, &vErr) || errors.As(err, &iErr) {
			status = 400
		} else if errors.As(err, &nErr) {
			status = 404
		}
		return LambdaResponse{StatusCode: status, Body: string(result)}, nil
	}

	// Return success response
	log.Info("Lambda function completed successfully")
	return LambdaResponse{
		StatusCode: 200,
		Body:       string(result),
	}, nil
}

// LambdaEvent represents the input event for Lambda
// ProcessingConfigPayload is used only for JSON decoding of the Lambda event
// to keep domain entities free of serialization concerns.
type ProcessingConfigPayload struct {
	FrameRate    float64 `json:"frame_rate"`
	OutputFormat string  `json:"output_format"`
}

type LambdaEvent struct {
	VideoKey      string                   `json:"video_key"`
	Configuration *ProcessingConfigPayload `json:"configuration,omitempty"`
}

// LambdaResponse represents the Lambda function response
type LambdaResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func toErrorBody(message, err string) string {
	m := map[string]any{
		"success": false,
		"message": message,
		"error":   err,
	}
	b, _ := json.Marshal(m)
	return string(b)
}
