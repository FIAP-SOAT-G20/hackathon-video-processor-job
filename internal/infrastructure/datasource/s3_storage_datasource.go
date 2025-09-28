package datasource

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

// S3StorageDataSource implements storage operations using AWS S3
type S3StorageDataSource struct {
	client          *s3.Client
	videoBucket     string
	processedBucket string
}

// NewS3StorageDataSource creates a new S3 storage datasource
func NewS3StorageDataSource(client *s3.Client, videoBucket, processedBucket string) port.StorageDataSource {
	return &S3StorageDataSource{
		client:          client,
		videoBucket:     videoBucket,
		processedBucket: processedBucket,
	}
}

// DownloadVideo downloads data from the video bucket in S3
func (ds *S3StorageDataSource) DownloadVideo(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := ds.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(ds.videoBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	return result.Body, nil
}

// UploadProcessedFile uploads data to the processed bucket in S3. Returns the object key that was uploaded.
func (ds *S3StorageDataSource) UploadProcessedFile(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error) {
	_, err := ds.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(ds.processedBucket),
		Key:           aws.String(key),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}
	return key, nil
}

// DeleteVideo deletes an object from the video bucket in S3
func (ds *S3StorageDataSource) DeleteVideo(ctx context.Context, key string) error {
	_, err := ds.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(ds.videoBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}
