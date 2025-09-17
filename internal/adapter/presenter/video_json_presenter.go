package presenter

import (
	"encoding/json"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

type videoJsonPresenter struct{}

func NewVideoJsonPresenter() port.Presenter {
	return &videoJsonPresenter{}
}

func (p *videoJsonPresenter) PresentProcessVideoOutput(output *dto.ProcessVideoOutput) []byte {
	response := VideoJsonResponse{
		Success:    output.Success,
		Message:    output.Message,
		OutputKey:  output.OutputKey,
		FrameCount: output.FrameCount,
		Hash:       output.Hash,
		Error:      output.Error,
	}

	result, _ := json.Marshal(response)
	return result
}

func (p *videoJsonPresenter) PresentError(err error) []byte {
	response := VideoJsonResponse{
		Success: false,
		Message: "Processing failed",
		Error:   err.Error(),
	}

	result, _ := json.Marshal(response)
	return result
}
