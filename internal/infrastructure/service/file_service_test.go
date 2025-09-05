package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalFileService(t *testing.T) {
	r := require.New(t)
	s := NewLocalFileService().(*LocalFileService)

	// Temp dir and file
	dir, err := s.CreateTempDir(context.Background(), "test_")
	r.NoError(err)
	r.DirExists(dir)
	defer s.DeleteDir(context.Background(), dir)

	f, err := s.CreateTempFile(context.Background(), "file_", ".txt")
	r.NoError(err)

	// Write and read
	err = s.WriteToFile(context.Background(), f, strings.NewReader("hello"))
	r.NoError(err)

	rc, err := s.ReadFile(context.Background(), f)
	r.NoError(err)
	rc.Close()

	// Size
	sz, err := s.GetFileSize(context.Background(), f)
	r.NoError(err)
	r.GreaterOrEqual(sz, int64(0))

	// List
	files, err := s.ListFiles(context.Background(), filepath.Dir(f), "*.txt")
	r.NoError(err)
	r.NotEmpty(files)

	// Delete
	r.NoError(s.DeleteFile(context.Background(), f))
	_, err = os.Stat(f)
	r.Error(err)
}
