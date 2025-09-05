package lambda

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
	"github.com/stretchr/testify/require"
)

type stubController struct {
	resp []byte
	err  error
}

func (s stubController) ProcessVideo(ctx context.Context, input dto.ProcessVideoInput) ([]byte, error) {
	return s.resp, s.err
}

func TestHandler(t *testing.T) {
	t.Run("BadJSON", func(t *testing.T) {
		r := require.New(t)
		h := NewHandler(stubController{}, logger.NewSlogLogger())
		_, err := h.Handle(context.Background(), []byte("{bad json"))
		r.NoError(err)
	})

	t.Run("MissingKey", func(t *testing.T) {
		r := require.New(t)
		h := NewHandler(stubController{}, logger.NewSlogLogger())
		body, err := h.Handle(context.Background(), []byte("{}"))
		r.NoError(err)
		// body should be LambdaResponse with 400
		b, _ := json.Marshal(body)
		var m map[string]any
		err = json.Unmarshal(b, &m)
		r.NoError(err)
		r.Equal(400, int(m["statusCode"].(float64)))
	})
}
