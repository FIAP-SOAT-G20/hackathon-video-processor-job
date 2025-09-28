package port

import (
	"context"
	"io"
)

// StorageDataSource defines the port for storage operations.
type StorageDataSource interface {
	DownloadVideo(ctx context.Context, key string) (io.ReadCloser, error)
	UploadProcessedFile(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error)
	DeleteVideo(ctx context.Context, key string) error
}
