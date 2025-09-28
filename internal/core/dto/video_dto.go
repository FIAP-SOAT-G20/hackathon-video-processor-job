package dto

// ProcessingConfigInput represents the input for video processing configuration
type ProcessingConfigInput struct {
	FrameRate    float64
	OutputFormat string
}

// ProcessVideoInput represents the input for video processing
type ProcessVideoInput struct {
	VideoKey      string
	UserId        string
	VideoId       string
	Configuration *ProcessingConfigInput
}

// ProcessVideoOutput represents the output of video processing
type ProcessVideoOutput struct {
	Success    bool
	Message    string
	OutputKey  string
	FrameCount int
	Hash       string
	Error      string
}
