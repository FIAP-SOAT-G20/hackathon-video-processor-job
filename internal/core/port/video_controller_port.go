package port

import (
	"context"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
)

type VideoController interface {
	ProcessVideo(ctx context.Context, input dto.ProcessVideoInput) ([]byte, error)
}
