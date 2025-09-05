package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	pmocks "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port/mocks"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
)

func TestVideoController(t *testing.T) {
	t.Run("ProcessVideo/success", func(t *testing.T) {
		r := require.New(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		uc := pmocks.NewMockVideoUseCase(ctrl)
		pr := pmocks.NewMockPresenter(ctrl)
		log := logger.NewSlogLogger()

		c := NewVideoController(uc, pr, log)

		input := dto.ProcessVideoInput{VideoKey: "foo.mp4"}
		out := &dto.ProcessVideoOutput{Success: true, Message: "ok", FrameCount: 5}
		prBytes := []byte(`{"success":true}`)

		uc.EXPECT().ProcessVideo(gomock.Any(), input).Return(out, nil)
		pr.EXPECT().PresentProcessVideoOutput(out).Return(prBytes)

		res, err := c.ProcessVideo(context.Background(), input)

		r.NoError(err)
		r.Equal(prBytes, res)
	})

	t.Run("ProcessVideo/error", func(t *testing.T) {
		r := require.New(t)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		uc := pmocks.NewMockVideoUseCase(ctrl)
		pr := pmocks.NewMockPresenter(ctrl)
		log := logger.NewSlogLogger()

		c := NewVideoController(uc, pr, log)

		input := dto.ProcessVideoInput{VideoKey: "bar.mp4"}
		errBoom := errors.New("boom")
		prBytes := []byte(`{"success":false}`)

		uc.EXPECT().ProcessVideo(gomock.Any(), input).Return(nil, errBoom)
		pr.EXPECT().PresentError(errBoom).Return(prBytes)

		res, err := c.ProcessVideo(context.Background(), input)
		r.Error(err)
		r.Equal(prBytes, res)
	})
}
