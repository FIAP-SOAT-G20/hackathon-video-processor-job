package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
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
		b, err := p.PresentProcessVideoOutput(out)
		require.NoError(t, err)
		var m map[string]any
		r.NoError(json.Unmarshal(b, &m))
		r.Equal(true, m["success"])
		r.Equal("ok", m["message"])
		r.Equal("processed/foo_frames.zip", m["output_key"])
		r.Equal(42, int(m["frame_count"].(float64)))
	})

	t.Run("PresentError", func(t *testing.T) {
		r := require.New(t)
		p := NewVideoJsonPresenter()
		b, err := p.PresentError(errors.New("boom"))
		require.NoError(t, err)
		var m map[string]any
		r.NoError(json.Unmarshal(b, &m))
		r.Equal(false, m["success"])
		r.Equal("Processing failed", m["message"])
		r.NotEmpty(m["error"])
	})
}
