package gateway

import (
	"context"
	"io"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

type videoGateway struct {
	storageDataSource StorageDataSource
}

type StorageDataSource interface {
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error)
	Delete(ctx context.Context, key string) error
}

func NewVideoGateway(storageDataSource StorageDataSource) port.VideoGateway {
	return &videoGateway{
		storageDataSource: storageDataSource,
	}
}

func (g *videoGateway) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return g.storageDataSource.Download(ctx, key)
}

func (g *videoGateway) Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error) {
	return g.storageDataSource.Upload(ctx, key, data, contentType, size)
}

func (g *videoGateway) Delete(ctx context.Context, key string) error {
	return g.storageDataSource.Delete(ctx, key)
}
