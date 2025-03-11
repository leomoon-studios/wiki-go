package i18n

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"wiki-go/internal/resources"
)

// Instead of using embed, we'll read the files directly from the filesystem
// This is a simpler approach that works well for development

// CopyLangsToStaticDir copies language files from embedded resources to data/static/langs
func CopyLangsToStaticDir(rootDir string) error {
	// Create static/langs directory if it doesn't exist
	staticLangsDir := filepath.Join(rootDir, "static", "langs")
	if err := os.MkdirAll(staticLangsDir, 0755); err != nil {
		return err
	}

	// Get embedded language files
	langFS := resources.GetLanguageFS()

	// Read all language files
	entries, err := fs.ReadDir(langFS, ".")
	if err != nil {
		// Fallback to direct filesystem access (for development)
		log.Printf("Warning: Embedded language files not available, trying filesystem: %v", err)
		return copyFromFilesystem(staticLangsDir)
	}

	// Copy each language file to the static directory
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process JSON files
		fileName := entry.Name()
		if filepath.Ext(fileName) != ".json" {
			continue
		}

		// Read embedded file
		fileData, err := fs.ReadFile(langFS, fileName)
		if err != nil {
			log.Printf("Warning: Failed to read embedded language file %s: %v", fileName, err)
			continue
		}

		// Write to static directory
		targetPath := filepath.Join(staticLangsDir, fileName)

		// Always overwrite existing files to ensure updates are applied
		if err := os.WriteFile(targetPath, fileData, 0644); err != nil {
			log.Printf("Warning: Failed to write language file %s: %v", targetPath, err)
			continue
		}

		log.Printf("Copied language file %s to %s", fileName, targetPath)
	}

	return nil
}

// Fallback function for development mode
func copyFromFilesystem(staticLangsDir string) error {
	// Try multiple possible locations for language files
	possiblePaths := []string{
		"internal/langs",
		"internal/resources/langs",
		"langs",
	}

	// Try each path
	var sourceLangsDir string
	var entries []os.DirEntry
	var readErr error
	foundPath := false

	for _, path := range possiblePaths {
		log.Printf("Trying to find language files in: %s", path)
		entries, readErr = os.ReadDir(path)
		if readErr == nil && len(entries) > 0 {
			// Make sure there are actual JSON files in the directory
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
					sourceLangsDir = path
					foundPath = true
					log.Printf("Found language files in: %s", path)
					break
				}
			}
			if foundPath {
				break
			}
		}
	}

	if !foundPath {
		return readErr
	}

	// Copy each language file to the static directory
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process JSON files
		fileName := entry.Name()
		if filepath.Ext(fileName) != ".json" {
			continue
		}

		// Read source file
		sourcePath := filepath.Join(sourceLangsDir, fileName)
		fileData, err := os.ReadFile(sourcePath)
		if err != nil {
			log.Printf("Warning: Failed to read language file %s: %v", fileName, err)
			continue
		}

		// Write to static directory
		targetPath := filepath.Join(staticLangsDir, fileName)

		// Always overwrite existing files to ensure updates are applied
		if err := os.WriteFile(targetPath, fileData, 0644); err != nil {
			log.Printf("Warning: Failed to write language file %s: %v", targetPath, err)
			continue
		}

		log.Printf("Copied language file %s to %s", fileName, targetPath)
	}

	return nil
}
