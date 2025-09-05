package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domain "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain"
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

	t.Run("ControllerValidationError => 400", func(t *testing.T) {
		r := require.New(t)
		h := NewHandler(stubController{resp: []byte(`{"success":false}`), err: domain.NewValidationError(errors.New("bad video"))}, logger.NewSlogLogger())
		body, err := h.Handle(context.Background(), []byte(`{"video_key":"foo.mp4"}`))
		r.NoError(err)
		b, _ := json.Marshal(body)
		var m map[string]any
		r.NoError(json.Unmarshal(b, &m))
		r.Equal(400, int(m["statusCode"].(float64)))
	})

	t.Run("ControllerNotFoundError => 404", func(t *testing.T) {
		r := require.New(t)
		h := NewHandler(stubController{resp: []byte(`{"success":false}`), err: domain.NewNotFoundError(domain.ErrNotFound)}, logger.NewSlogLogger())
		body, err := h.Handle(context.Background(), []byte(`{"video_key":"foo.mp4"}`))
		r.NoError(err)
		b, _ := json.Marshal(body)
		var m map[string]any
		r.NoError(json.Unmarshal(b, &m))
		r.Equal(404, int(m["statusCode"].(float64)))
	})

	t.Run("ControllerInternalError => 500", func(t *testing.T) {
		r := require.New(t)
		h := NewHandler(stubController{resp: []byte(`{"success":false}`), err: domain.NewInternalError(errors.New("boom"))}, logger.NewSlogLogger())
		body, err := h.Handle(context.Background(), []byte(`{"video_key":"foo.mp4"}`))
		r.NoError(err)
		b, _ := json.Marshal(body)
		var m map[string]any
		r.NoError(json.Unmarshal(b, &m))
		r.Equal(500, int(m["statusCode"].(float64)))
	})
}
