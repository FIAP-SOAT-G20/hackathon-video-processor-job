package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

const (
	// DefaultJPEGQuality defines the quality level for JPEG output (1-31, lower = better quality)
	// Value 2 provides high quality with reasonable file size
	DefaultJPEGQuality = "2"
)

type FFmpegService struct {
	fileManager port.FileManager
}

func NewFFmpegService(fileManager port.FileManager) port.VideoProcessor {
	return &FFmpegService{
		fileManager: fileManager,
	}
}

// ProcessVideo processes video and extracts frames using FFmpeg
func (s *FFmpegService) ProcessVideo(ctx context.Context, videoPath string, frameRate float64, outputFormat string) (int, string, error) {
	// Create temporary directory for frames
	tempDir, err := s.fileManager.CreateTempDir(ctx, "frames_")
	if err != nil {
		return 0, "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = s.fileManager.DeleteDir(ctx, tempDir)
	}()

	// Extract frames
	framePaths, err := s.extractFrames(ctx, videoPath, frameRate, outputFormat, tempDir)
	if err != nil {
		return 0, "", fmt.Errorf("failed to extract frames: %w", err)
	}

	if len(framePaths) == 0 {
		return 0, "", fmt.Errorf("no frames extracted from video")
	}

	// Create ZIP file
	zipPath, err := s.fileManager.CreateTempFile(ctx, "frames_", ".zip")
	if err != nil {
		return 0, "", fmt.Errorf("failed to create temp zip file: %w", err)
	}

	if err := s.createZipFromFiles(framePaths, zipPath); err != nil {
		return 0, "", fmt.Errorf("failed to create zip file: %w", err)
	}

	return len(framePaths), zipPath, nil
}

// ValidateVideo checks if video file is valid and can be processed
func (s *FFmpegService) ValidateVideo(ctx context.Context, videoPath string) error {
	// Check if file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("video file does not exist: %s", videoPath)
	}

	// Use FFprobe to validate video
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-count_packets",
		"-show_entries", "stream=nb_read_packets",
		"-of", "csv=p=0",
		videoPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("video validation failed: %w\nffprobe output:\n%s", err, string(output))
	}

	// Check if we got packet count (indicates valid video stream)
	packetCount := strings.TrimSpace(string(output))
	// Empty output indicates no video stream, but 0 packets can be valid for very short videos
	if packetCount == "" {
		return fmt.Errorf("video file contains no valid video stream")
	}

	return nil
}

func (s *FFmpegService) extractFrames(ctx context.Context, videoPath string, frameRate float64, outputFormat string, outputDir string) ([]string, error) {
	framePattern := filepath.Join(outputDir, fmt.Sprintf("frame_%%04d.%s", outputFormat))

	var args []string
	if outputFormat == "jpg" {
		args = []string{
			"-nostdin",
			"-loglevel", "error",
			"-y",
			"-i", videoPath,
			"-map", "0:v:0",
			"-an",
			"-vf", fmt.Sprintf("fps=%g", frameRate),
			"-start_number", "0",
			"-f", "image2",
			"-vcodec", "mjpeg",
			"-q:v", DefaultJPEGQuality,
			"-frame_pts", "1",
			framePattern,
		}
	} else { // png
		args = []string{
			"-nostdin",
			"-loglevel", "error",
			"-y",
			"-i", videoPath,
			"-map", "0:v:0",
			"-an",
			"-vf", fmt.Sprintf("fps=%g", frameRate),
			"-start_number", "0",
			"-f", "image2",
			"-vcodec", "png",
			"-frame_pts", "1",
			framePattern,
		}
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	pattern := fmt.Sprintf("*.%s", outputFormat)
	framePaths, err := s.fileManager.ListFiles(ctx, outputDir, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list frame files: %w", err)
	}

	return framePaths, nil
}
func (s *FFmpegService) createZipFromFiles(files []string, outputPath string) error {
	// Create output file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func() {
		_ = zipFile.Close()
	}()

	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		_ = zipWriter.Close()
	}()

	// Add each file to the ZIP
	for _, filePath := range files {
		if err := s.addFileToZip(zipWriter, filePath); err != nil {
			return fmt.Errorf("failed to add file %s to zip: %w", filePath, err)
		}
	}

	return nil
}

func (s *FFmpegService) addFileToZip(zipWriter *zip.Writer, filePath string) error {
	// Open source file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create ZIP file header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}

	// Use filename without path as the name in ZIP
	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	// Create writer for this file in ZIP
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	// Copy file content to ZIP
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to write file to zip: %w", err)
	}

	return nil
}
