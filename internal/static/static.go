package static

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"wiki-go/internal/resources"
)

// List of static files to copy from resources to data/static
var staticFiles = []string{
	"favicon.ico",
	"favicon.png",
	"favicon.svg",
}

// EnsureStaticAssetsExist copies default static assets to data/static directory if they don't exist
func EnsureStaticAssetsExist(dataDir string) error {
	// Create the static directory path
	staticDir := filepath.Join(dataDir, "static")

	// Ensure the static directory exists
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %w", err)
	}

	// Get the embedded filesystem
	fsys := resources.GetFileSystem()

	// Copy each static file if it doesn't exist
	for _, filename := range staticFiles {
		destPath := filepath.Join(staticDir, filename)

		// Check if file already exists in data/static
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			// File doesn't exist, copy from embedded resources

			// Open source file from embedded resources
			srcFile, err := fsys.Open(filename)
			if err != nil {
				log.Printf("Warning: Cannot find embedded resource %s: %v", filename, err)
				continue
			}
			defer srcFile.Close()

			// Create destination file
			destFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
			}
			defer destFile.Close()

			// Copy content
			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return fmt.Errorf("failed to copy file %s: %w", filename, err)
			}

			log.Printf("Copied default static file to %s", destPath)
		}
	}

	return nil
}

// GetStaticFilePath returns the path to a static file, checking data/static first, then falling back to internal resources
func GetStaticFilePath(dataDir, filename string) string {
	// First check if the file exists in data/static
	customPath := filepath.Join(dataDir, "static", filename)
	if _, err := os.Stat(customPath); err == nil {
		// File exists in data/static, use it
		return customPath
	}

	// For embedded resources, we don't return a path since they're not on the filesystem
	// Instead, we'll handle this in the ServeStaticFile function
	// This is just a fallback for compatibility
	return filepath.Join(dataDir, "static", filename)
}

// ServeStaticFile writes a static file to the given writer, checking data/static first
func ServeStaticFile(w io.Writer, dataDir, filename string) error {
	// First check if the file exists in data/static
	customPath := filepath.Join(dataDir, "static", filename)
	if _, err := os.Stat(customPath); err == nil {
		// File exists in data/static, serve it directly
		file, err := os.Open(customPath)
		if err != nil {
			return fmt.Errorf("failed to open static file %s: %w", customPath, err)
		}
		defer file.Close()

		_, err = io.Copy(w, file)
		return err
	}

	// Fall back to embedded resources
	fsys := resources.GetFileSystem()
	file, err := fsys.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open embedded resource %s: %w", filename, err)
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}
