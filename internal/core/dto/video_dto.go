package dto

import "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain/entity"

// ProcessVideoInput represents the input for video processing
type ProcessVideoInput struct {
	VideoKey      string
	Configuration *entity.ProcessingConfig
}

// ProcessVideoOutput represents the output of video processing
type ProcessVideoOutput struct {
	Success    bool
	Message    string
	OutputKey  string
	FrameCount int
	Error      string
}
