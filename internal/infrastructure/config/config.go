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
		Bucket          string
		ProcessedBucket string
		ExportFormat    string
		ExportFPS       float64
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
	config.AWS.Region = getEnv("K8S_JOB_ENV_AWS_REGION", "us-east-1")
	config.AWS.AccessKey = getEnv("K8S_JOB_ENV_AWS_ACCESS_KEY_ID", "")
	config.AWS.SecretAccessKey = getEnv("K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY", "")
	config.AWS.SessionToken = getEnv("K8S_JOB_ENV_AWS_SESSION_TOKEN", "")

	// Video Configuration
	config.Video.Key = getEnv("K8S_JOB_ENV_VIDEO_KEY", "")
	config.Video.Bucket = getEnv("K8S_JOB_ENV_VIDEO_BUCKET", "video-processor-raw-videos")
	config.Video.ProcessedBucket = getEnv("K8S_JOB_ENV_PROCESSED_BUCKET", "video-processor-processed-images")
	config.Video.ExportFormat = getEnv("K8S_JOB_ENV_VIDEO_EXPORT_FORMAT", "jpg")

	// Parse frame rate
	frameRateStr := getEnv("K8S_JOB_ENV_VIDEO_EXPORT_FPS", "1.0")
	frameRate, err := strconv.ParseFloat(frameRateStr, 64)
	if err != nil {
		log.Printf("Warning: Invalid K8S_JOB_ENV_VIDEO_EXPORT_FPS '%s', using default 1.0: %v", frameRateStr, err)
		frameRate = 1.0
	}
	config.Video.ExportFPS = frameRate

	return config
}

// ValidateRequiredFields validates that all required configuration fields are set
func (c *Config) ValidateRequiredFields() error {
	var missingFields []string

	if c.AWS.AccessKey == "" {
		missingFields = append(missingFields, "K8S_JOB_ENV_AWS_ACCESS_KEY_ID")
	}
	if c.AWS.SecretAccessKey == "" {
		missingFields = append(missingFields, "K8S_JOB_ENV_AWS_SECRET_ACCESS_KEY")
	}
	if c.AWS.SessionToken == "" {
		missingFields = append(missingFields, "K8S_JOB_ENV_AWS_SESSION_TOKEN")
	}
	if c.Video.Key == "" {
		missingFields = append(missingFields, "K8S_JOB_ENV_VIDEO_KEY")
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
