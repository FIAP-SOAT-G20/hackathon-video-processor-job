package datasource

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/adapter/gateway"
)

// S3StorageDataSource implements storage operations using AWS S3
type S3StorageDataSource struct {
	client          *s3.Client
	videoBucket     string
	processedBucket string
}

// NewS3StorageDataSource creates a new S3 storage datasource
func NewS3StorageDataSource(client *s3.Client, videoBucket, processedBucket string) gateway.StorageDataSource {
	return &S3StorageDataSource{
		client:          client,
		videoBucket:     videoBucket,
		processedBucket: processedBucket,
	}
}

// Download downloads data from S3
func (ds *S3StorageDataSource) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	bucket := ds.videoBucket
	if strings.Contains(key, "processed/") {
		bucket = ds.processedBucket
	}

	result, err := ds.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// Upload uploads data to S3
func (ds *S3StorageDataSource) Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) (string, error) {
	bucket := ds.processedBucket
	if strings.Contains(key, "raw/") || strings.Contains(key, "video/") {
		bucket = ds.videoBucket
	}

	_, err := ds.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return S3 URL
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	return url, nil
}

// Delete deletes an object from S3
func (ds *S3StorageDataSource) Delete(ctx context.Context, key string) error {
	bucket := ds.videoBucket
	if strings.Contains(key, "processed/") {
		bucket = ds.processedBucket
	}

	_, err := ds.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}
