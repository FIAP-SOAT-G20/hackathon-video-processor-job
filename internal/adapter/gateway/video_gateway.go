package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

type videoGateway struct {
	storageDataSource port.StorageDataSource
	messageBroker     port.MessageBroker
}

func NewVideoGateway(storageDataSource port.StorageDataSource, messageBroker port.MessageBroker) port.VideoGateway {
	return &videoGateway{
		storageDataSource: storageDataSource,
		messageBroker:     messageBroker,
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

func (g *videoGateway) UpdateStatus(ctx context.Context, videoId string, userId string, videoHash string, status string) error {
	videoIdUint, err := strconv.ParseUint(videoId, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse video id: %w", err)
	}
	userIdUint, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse user id: %w", err)
	}
	body := map[string]any{
		"video_id": videoIdUint,
		"user_id":  userIdUint,
		"hash":     videoHash,
		"status":   status,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	err = g.messageBroker.PublishMessage(ctx, jsonBody)
	if err != nil {
		return fmt.Errorf("failed to publish status update: %w", err)
	}

	return nil
}
