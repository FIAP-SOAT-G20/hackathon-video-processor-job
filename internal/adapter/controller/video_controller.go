package controller

import (
	"context"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
)

type videoController struct {
	videoUseCase port.VideoUseCase
	presenter    port.Presenter
	logger       logger.Logger
}

func NewVideoController(videoUseCase port.VideoUseCase, presenter port.Presenter, logger logger.Logger) port.VideoController {
	return &videoController{
		videoUseCase: videoUseCase,
		presenter:    presenter,
		logger:       logger,
	}
}

func (c *videoController) ProcessVideo(ctx context.Context, input dto.ProcessVideoInput) ([]byte, error) {
	log := c.logger.WithContext(ctx).With("video_key", input.VideoKey)
	log.Info("Controller received video processing request")

	output, err := c.videoUseCase.ProcessVideo(ctx, input)
	if err != nil {
		log.Error("Video processing failed", "error", err)
		b, pErr := c.presenter.PresentError(err)
		if pErr != nil {
			log.Error("Failed to marshal error response", "error", pErr)
			return nil, pErr
		}
		return b, err
	}

	log.Info("Video processing completed successfully", "success", output.Success, "frame_count", output.FrameCount)
	b, pErr := c.presenter.PresentProcessVideoOutput(output)
	if pErr != nil {
		log.Error("Failed to marshal success response", "error", pErr)
		return nil, pErr
	}
	return b, nil
}
