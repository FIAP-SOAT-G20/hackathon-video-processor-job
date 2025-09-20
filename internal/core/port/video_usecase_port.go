package port

import (
	"context"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
)

type VideoUseCase interface {
	ProcessVideo(ctx context.Context, input dto.ProcessVideoInput) (*dto.ProcessVideoOutput, error)
}
