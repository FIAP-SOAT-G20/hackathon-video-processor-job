package port

import "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"

type Presenter interface {
	PresentProcessVideoOutput(output *dto.ProcessVideoOutput) ([]byte, error)
	PresentError(err error) ([]byte, error)
}
