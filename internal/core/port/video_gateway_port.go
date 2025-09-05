package port

import (
	"context"
	"io"
)

type VideoGateway interface {
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error)
	Delete(ctx context.Context, key string) error
}

type VideoProcessor interface {
	ProcessVideo(ctx context.Context, videoPath string, frameRate float64) ([]string, int, string, error)
	ValidateVideo(ctx context.Context, videoPath string) error
}

type FileManager interface {
	CreateTempFile(ctx context.Context, prefix, suffix string) (string, error)
	CreateTempDir(ctx context.Context, prefix string) (string, error)
	WriteToFile(ctx context.Context, filePath string, data io.Reader) error
	ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, filePath string) error
	DeleteDir(ctx context.Context, dirPath string) error
	ListFiles(ctx context.Context, dirPath, pattern string) ([]string, error)
	GetFileSize(ctx context.Context, filePath string) (int64, error)
}
