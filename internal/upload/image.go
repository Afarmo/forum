package upload

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"uuid"

	"github.com/Afarmo/forum/internal/apperrors"
)

const (
	MaxImageSize   = 5 << 20
	MaxRequestSize = MaxImageSize + 1<<20 // Extra 1 MiB for multipart form data
)

// SaveImage validates and saves an uploaded image to the specified directory.
// It generates a unique filename and returns it on success.
// Only JPEG, PNG, and WebP images are supported.
func SaveImage(file multipart.File, size int64, dir string) (string, error) {
	if size > MaxImageSize {
		return "", apperrors.ErrImageTooLarge
	}

	// Read the first 512 bytes to detect the actual file type.
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	// Validate the MIME type instead of trusting the file extension.
	var ext string
	switch http.DetectContentType(buf[:n]) {
	case "image/jpeg":
		ext = ".jpeg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", apperrors.ErrInvalidImage
	}

	// Reset the reader so the entire image gets saved.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Create the destination directory if it doesn't exist.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Generate a unique filename
	filename := uuid.New().String() + ext
	path := filepath.Join(dir, filename)

	destination, err := os.Create(path)
	if err != nil {
		return "", err
	}

	// Copy at most 5 MiB + 1 byte to detect oversized files.
	written, copyErr := io.Copy(destination, io.LimitReader(file, MaxImageSize+1))
	closeErr := destination.Close()

	// Remove incomplete or oversized files if saving fails.
	if copyErr != nil || closeErr != nil || written > MaxImageSize {
		os.Remove(path)

		if written > MaxImageSize {
			return "", apperrors.ErrImageTooLarge
		}

		return "", errors.Join(copyErr, closeErr)
	}

	return filename, nil
}
