package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

// LocalFileService implements file operations for Lambda execution environment
type LocalFileService struct {
	tempDir string
}

// NewLocalFileService creates a new local file service
func NewLocalFileService() port.FileManager {
	tempDir := "/tmp" // Lambda temp directory
	if os.Getenv("LAMBDA_RUNTIME_DIR") == "" {
		// Running locally, use system temp
		tempDir = os.TempDir()
	}

	return &LocalFileService{
		tempDir: tempDir,
	}
}

// CreateTempFile creates a temporary file with given prefix and suffix
func (s *LocalFileService) CreateTempFile(ctx context.Context, prefix, suffix string) (string, error) {
	file, err := os.CreateTemp(s.tempDir, prefix+"*"+suffix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	path := file.Name()
	file.Close()

	return path, nil
}

// CreateTempDir creates a temporary directory with given prefix
func (s *LocalFileService) CreateTempDir(ctx context.Context, prefix string) (string, error) {
	dir, err := os.MkdirTemp(s.tempDir, prefix+"*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	return dir, nil
}

// WriteToFile writes data to a file
func (s *LocalFileService) WriteToFile(ctx context.Context, filePath string, data io.Reader) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// ReadFile reads data from a file
func (s *LocalFileService) ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// DeleteFile deletes a file
func (s *LocalFileService) DeleteFile(ctx context.Context, filePath string) error {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteDir deletes a directory and all its contents
func (s *LocalFileService) DeleteDir(ctx context.Context, dirPath string) error {
	if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// ListFiles lists files in directory matching pattern
func (s *LocalFileService) ListFiles(ctx context.Context, dirPath, pattern string) ([]string, error) {
	searchPattern := filepath.Join(dirPath, pattern)
	matches, err := filepath.Glob(searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Filter to only include files (not directories)
	var files []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && !info.IsDir() {
			files = append(files, match)
		}
	}

	return files, nil
}

// GetFileSize returns the size of a file
func (s *LocalFileService) GetFileSize(ctx context.Context, filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return info.Size(), nil
}
