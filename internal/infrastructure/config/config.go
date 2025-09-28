package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// AWS Settings
	AWS struct {
		Region          string
		AccessKey       string
		SecretAccessKey string
		SessionToken    string
	}

	// Video Processing Settings
	Video struct {
		Key             string
		Id              string
		UserId          string
		Bucket          string
		ProcessedBucket string
		ExportFormat    string
		ExportFPS       float64
		SnsTopic        string
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	config := &Config{}

	// AWS Configuration
	config.AWS.Region = getEnv("AWS_REGION", "us-east-1")
	config.AWS.AccessKey = getEnv("AWS_ACCESS_KEY_ID", "")
	config.AWS.SecretAccessKey = getEnv("AWS_SECRET_ACCESS_KEY", "")
	config.AWS.SessionToken = getEnv("AWS_SESSION_TOKEN", "")

	// Video Configuration
	config.Video.Key = getEnv("VIDEO_KEY", "")
	config.Video.Id = getEnv("VIDEO_ID", "1")
	config.Video.UserId = getEnv("VIDEO_USER_ID", "1")
	config.Video.Bucket = getEnv("VIDEO_BUCKET", "video-processor-raw-videos")
	config.Video.ProcessedBucket = getEnv("PROCESSED_BUCKET", "video-processor-processed-images")
	config.Video.ExportFormat = getEnv("VIDEO_EXPORT_FORMAT", "jpg")
	config.Video.SnsTopic = getEnv("SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:905417995957:video-status-updated")
	// Parse frame rate
	frameRateStr := getEnv("VIDEO_EXPORT_FPS", "1.0")
	frameRate, err := strconv.ParseFloat(frameRateStr, 64)
	if err != nil {
		log.Printf("Warning: Invalid VIDEO_EXPORT_FPS '%s', using default 1.0: %v", frameRateStr, err)
		frameRate = 1.0
	}
	config.Video.ExportFPS = frameRate

	return config
}

// ValidateRequiredFields validates that all required configuration fields are set
func (c *Config) ValidateRequiredFields() error {
	var missingFields []string

	if c.AWS.AccessKey == "" {
		missingFields = append(missingFields, "AWS_ACCESS_KEY_ID")
	}
	if c.AWS.SecretAccessKey == "" {
		missingFields = append(missingFields, "AWS_SECRET_ACCESS_KEY")
	}
	if c.AWS.SessionToken == "" {
		missingFields = append(missingFields, "AWS_SESSION_TOKEN")
	}
	if c.Video.Key == "" {
		missingFields = append(missingFields, "VIDEO_KEY")
	}

	if len(missingFields) > 0 {
		return &ConfigValidationError{MissingFields: missingFields}
	}

	return nil
}

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	MissingFields []string
}

func (e *ConfigValidationError) Error() string {
	return "missing required environment variables"
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
