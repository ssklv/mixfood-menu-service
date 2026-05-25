package infrastructure

import (
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type FileStorage interface {
	SaveFile(file io.Reader, filename string) (string, error)
}

type localFileStorage struct {
	uploadDir string
}

func NewLocalFileStorage(dir string) (FileStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &localFileStorage{uploadDir: dir}, nil
}

func (s *localFileStorage) SaveFile(file io.Reader, filename string) (string, error) {
	ext := filepath.Ext(filename)
	newFilename := uuid.New().String() + ext
	path := filepath.Join(s.uploadDir, newFilename)

	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/uploads/" + newFilename, nil
}
