package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/domain"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/dto"
	pmocks "github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port/mocks"
	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/infrastructure/logger"
)

func TestVideoUseCase(t *testing.T) {
	// Group tests related to video use case behavior
	t.Run("GenerateOutputKey", func(t *testing.T) {
		uc := &videoUseCase{}

		cases := []struct {
			name string
			in   string
			want string
		}{
			{"simple", "video.mp4", "processed/video_frames.zip"},
			{"nested_path", "path/to/file.avi", "processed/path/to/file_frames.zip"},
			{"no_ext", "noext", "processed/noext_frames.zip"},
			{"multi_dot", "multi.dot.name.mp4", "processed/multi.dot.name_frames.zip"},
		}

		for _, tc := range cases {
			tc := tc // capture range var
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel() // safe: pure function
				got := uc.generateOutputKey(tc.in)
				require.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("ProcessVideo", func(t *testing.T) {
		t.Run("success_default_config", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vg := pmocks.NewMockVideoGateway(ctrl)
			vp := pmocks.NewMockVideoProcessor(ctrl)
			fm := pmocks.NewMockFileManager(ctrl)
			log := logger.NewSlogLogger()

			uc := NewVideoUseCase(vg, vp, fm, log)

			videoKey := "folder/foo.mp4"
			localPath := "/tmp/video123.mp4"
			zipPath := "/tmp/frames.zip"

			// Download to local
			fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
			vg.EXPECT().Download(gomock.Any(), videoKey).Return(io.NopCloser(strings.NewReader("data")), nil)
			fm.EXPECT().WriteToFile(gomock.Any(), localPath, gomock.Any()).Return(nil)

			// Validate and process (defaults: 1.0, png)
			vp.EXPECT().ValidateVideo(gomock.Any(), localPath).Return(nil)
			vp.EXPECT().ProcessVideo(gomock.Any(), localPath, 1.0, "png").Return([]string{"f1.png"}, 1, zipPath, nil)

			// Upload result
			fm.EXPECT().ReadFile(gomock.Any(), zipPath).Return(io.NopCloser(bytes.NewBufferString("zipdata")), nil)
			fm.EXPECT().GetFileSize(gomock.Any(), zipPath).Return(int64(7), nil)
			vg.EXPECT().Upload(gomock.Any(), "processed/folder/foo_frames.zip", gomock.Any(), "application/zip", int64(7)).Return("processed/folder/foo_frames.zip", nil)

			// Delete original video and temp files
			vg.EXPECT().Delete(gomock.Any(), videoKey).Return(nil)
			fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)
			fm.EXPECT().DeleteFile(gomock.Any(), zipPath).Return(nil)

			out, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			require.NoError(t, err)
			require.True(t, out.Success)
			require.Equal(t, 1, out.FrameCount)
			require.Equal(t, "processed/folder/foo_frames.zip", out.OutputKey)
		})

		t.Run("validation_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vg := pmocks.NewMockVideoGateway(ctrl)
			vp := pmocks.NewMockVideoProcessor(ctrl)
			fm := pmocks.NewMockFileManager(ctrl)
			log := logger.NewSlogLogger()

			uc := NewVideoUseCase(vg, vp, fm, log)

			videoKey := "bad.mp4"
			localPath := "/tmp/video_bad.mp4"

			fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
			vg.EXPECT().Download(gomock.Any(), videoKey).Return(io.NopCloser(strings.NewReader("data")), nil)
			fm.EXPECT().WriteToFile(gomock.Any(), localPath, gomock.Any()).Return(nil)
			vp.EXPECT().ValidateVideo(gomock.Any(), localPath).Return(errors.New("boom"))
			fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)

			out, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			require.Error(t, err)
			require.False(t, out.Success)
			var vErr *domain.ValidationError
			require.ErrorAs(t, err, &vErr)
		})

		// ensure download errors are wrapped as InternalError
		t.Run("download_error_internal", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vg := pmocks.NewMockVideoGateway(ctrl)
			vp := pmocks.NewMockVideoProcessor(ctrl)
			fm := pmocks.NewMockFileManager(ctrl)
			log := logger.NewSlogLogger()

			uc := NewVideoUseCase(vg, vp, fm, log)

			videoKey := "folder/foo.mp4"
			localPath := "/tmp/video123.mp4"

			fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
			vg.EXPECT().Download(gomock.Any(), videoKey).Return(nil, errors.New("download failed"))

			_, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			var iErr *domain.InternalError
			require.ErrorAs(t, err, &iErr)
		})

		// zero frames should produce InvalidInputError and cleanup temp files
		t.Run("zero_frames_invalid_input", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vg := pmocks.NewMockVideoGateway(ctrl)
			vp := pmocks.NewMockVideoProcessor(ctrl)
			fm := pmocks.NewMockFileManager(ctrl)
			log := logger.NewSlogLogger()

			uc := NewVideoUseCase(vg, vp, fm, log)

			videoKey := "folder/empty.mp4"
			localPath := "/tmp/empty.mp4"
			zipPath := "/tmp/frames0.zip"

			fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
			vg.EXPECT().Download(gomock.Any(), videoKey).Return(io.NopCloser(strings.NewReader("data")), nil)
			fm.EXPECT().WriteToFile(gomock.Any(), localPath, gomock.Any()).Return(nil)
			vp.EXPECT().ValidateVideo(gomock.Any(), localPath).Return(nil)
			vp.EXPECT().ProcessVideo(gomock.Any(), localPath, 1.0, "png").Return([]string{}, 0, zipPath, nil)

			// defers should cleanup these files when error occurs
			fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)
			fm.EXPECT().DeleteFile(gomock.Any(), zipPath).Return(nil)

			_, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			var invErr *domain.InvalidInputError
			require.ErrorAs(t, err, &invErr)
		})

		t.Run("download_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			vg := pmocks.NewMockVideoGateway(ctrl)
			vp := pmocks.NewMockVideoProcessor(ctrl)
			fm := pmocks.NewMockFileManager(ctrl)
			log := logger.NewSlogLogger()

			uc := NewVideoUseCase(vg, vp, fm, log)

			videoKey := "missing.mp4"
			localPath := "/tmp/missing.mp4"

			fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
			vg.EXPECT().Download(gomock.Any(), videoKey).Return(nil, domain.NewNotFoundError(domain.ErrNotFound))

			out, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			var nf *domain.NotFoundError
			require.ErrorAs(t, err, &nf)
			require.False(t, out.Success)
		})
	})
}
