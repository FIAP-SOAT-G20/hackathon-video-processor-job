package presenter

type VideoJsonResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputKey  string `json:"output_key,omitempty"`
	FrameCount int    `json:"frame_count,omitempty"`
	Error      string `json:"error,omitempty"`
}
