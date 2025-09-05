package presenter

import (
	"encoding/json"
	"testing"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/stretchr/testify/require"
)

func TestVideoJsonPresenter(t *testing.T) {
	t.Run("PresentProcessVideoOutput_Success", func(t *testing.T) {
		r := require.New(t)
		p := NewVideoJsonPresenter()
		out := &dto.ProcessVideoOutput{
			Success:    true,
			Message:    "ok",
			OutputKey:  "processed/foo_frames.zip",
			FrameCount: 42,
		}
		b := p.PresentProcessVideoOutput(out)
		var m map[string]any
		err := json.Unmarshal(b, &m)
		r.NoError(err)
		r.Equal(true, m["success"])
		r.Equal("ok", m["message"])
		r.Equal("processed/foo_frames.zip", m["output_key"])
		r.Equal(42, int(m["frame_count"].(float64)))
	})

	t.Run("PresentError", func(t *testing.T) {
		r := require.New(t)
		p := NewVideoJsonPresenter()
		b := p.PresentError(assertErr{})
		var m map[string]any
		err := json.Unmarshal(b, &m)
		r.NoError(err)
		r.Equal(false, m["success"])
		r.Equal("Processing failed", m["message"])
		r.NotEmpty(m["error"])
	})
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }
