package gateway

import (
	"context"
	"io"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

type videoGateway struct {
	storageDataSource port.StorageDataSource
}

func NewVideoGateway(storageDataSource port.StorageDataSource) port.VideoGateway {
	return &videoGateway{
		storageDataSource: storageDataSource,
	}
}

func (g *videoGateway) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return g.storageDataSource.DownloadVideo(ctx, key)
}

func (g *videoGateway) Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error) {
	return g.storageDataSource.UploadProcessedFile(ctx, key, data, contentType, size)
}

func (g *videoGateway) Delete(ctx context.Context, key string) error {
	return g.storageDataSource.DeleteVideo(ctx, key)
}
