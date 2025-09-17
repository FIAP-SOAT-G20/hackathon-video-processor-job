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
			testVideoData := "test video data"
			// Hash will be calculated dynamically from the actual content

			// Download to local and generate hash
			fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
			vg.EXPECT().Download(gomock.Any(), videoKey).Return(io.NopCloser(strings.NewReader(testVideoData)), nil)
			fm.EXPECT().WriteToFile(gomock.Any(), localPath, gomock.Any()).Return(nil)

			// Validate and process (defaults: 1.0, png)
			vp.EXPECT().ValidateVideo(gomock.Any(), localPath).Return(nil)
			vp.EXPECT().ProcessVideo(gomock.Any(), localPath, 1.0, "png").Return([]string{"f1.png"}, 1, zipPath, nil)

			// Upload result using hash - mock returns any key that is passed
			fm.EXPECT().ReadFile(gomock.Any(), zipPath).Return(io.NopCloser(bytes.NewBufferString("zipdata")), nil)
			fm.EXPECT().GetFileSize(gomock.Any(), zipPath).Return(int64(7), nil)
			vg.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), "application/zip", int64(7)).DoAndReturn(
				func(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error) {
					return key, nil // returns the same key that was passed
				})

			// Delete original video and temp files
			vg.EXPECT().Delete(gomock.Any(), videoKey).Return(nil)
			fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)
			fm.EXPECT().DeleteFile(gomock.Any(), zipPath).Return(nil)

			out, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			require.NoError(t, err)
			require.True(t, out.Success)
			require.Equal(t, 1, out.FrameCount)
			// Verify that the hash was generated and used in OutputKey
			require.NotEmpty(t, out.Hash)
			require.Equal(t, "processed/"+out.Hash+".zip", out.OutputKey)
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
			// File cleanup is handled internally by downloadAndValidateVideo on error
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
			// Cleanup temp file on download error
			fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)

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
			// Cleanup temp file on download error
			fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)

			out, err := uc.ProcessVideo(context.Background(), dto.ProcessVideoInput{VideoKey: videoKey})
			var nf *domain.NotFoundError
			require.ErrorAs(t, err, &nf)
			require.False(t, out.Success)
		})
	})

	t.Run("ProcessVideo_custom_config_sanitization_frameRate_and_format_lowercase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		vg := pmocks.NewMockVideoGateway(ctrl)
		vp := pmocks.NewMockVideoProcessor(ctrl)
		fm := pmocks.NewMockFileManager(ctrl)
		log := logger.NewSlogLogger()

		uc := NewVideoUseCase(vg, vp, fm, log)

		videoKey := "vid.mp4"
		localPath := "/tmp/vid.mp4"
		zipPath := "/tmp/zip1.zip"
		testVideoData := "custom test video data"

		fm.EXPECT().CreateTempFile(gomock.Any(), "video_", ".mp4").Return(localPath, nil)
		vg.EXPECT().Download(gomock.Any(), videoKey).Return(io.NopCloser(strings.NewReader(testVideoData)), nil)
		fm.EXPECT().WriteToFile(gomock.Any(), localPath, gomock.Any()).Return(nil)
		vp.EXPECT().ValidateVideo(gomock.Any(), localPath).Return(nil)
		// input has frame_rate=0 (sanitize to 1.0) and output_format="JPG" (lowercase to "jpg")
		vp.EXPECT().ProcessVideo(gomock.Any(), localPath, 1.0, "jpg").Return([]string{"f1.jpg"}, 1, zipPath, nil)
		fm.EXPECT().ReadFile(gomock.Any(), zipPath).Return(io.NopCloser(bytes.NewBufferString("zip")), nil)
		fm.EXPECT().GetFileSize(gomock.Any(), zipPath).Return(int64(3), nil)
		vg.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), "application/zip", int64(3)).DoAndReturn(
			func(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (string, error) {
				return key, nil // returns the same key that was passed
			})
		vg.EXPECT().Delete(gomock.Any(), videoKey).Return(nil)
		fm.EXPECT().DeleteFile(gomock.Any(), localPath).Return(nil)
		fm.EXPECT().DeleteFile(gomock.Any(), zipPath).Return(nil)

		in := dto.ProcessVideoInput{VideoKey: videoKey, Configuration: &dto.ProcessingConfigInput{FrameRate: 0, OutputFormat: "JPG"}}
		out, err := uc.ProcessVideo(context.Background(), in)
		require.NoError(t, err)
		require.True(t, out.Success)
		require.Equal(t, 1, out.FrameCount)
		// Verify that the hash was generated and used in OutputKey
		require.NotEmpty(t, out.Hash)
		require.Equal(t, "processed/"+out.Hash+".zip", out.OutputKey)
		// Test that sanitization worked: frame_rate=0 -> 1.0, JPG -> jpg
	})
}
