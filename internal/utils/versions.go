package utils

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CleanupOldVersions removes old versions if the number of versions exceeds maxVersions
func CleanupOldVersions(versionDir string, maxVersions int) {
	// If maxVersions is 0 or negative, keep all versions
	if maxVersions <= 0 {
		return
	}

	// Read all files in the versions directory
	files, err := os.ReadDir(versionDir)
	if err != nil {
		log.Printf("Error reading versions directory: %v", err)
		return
	}

	// Filter and collect version files
	var versions []string
	for _, file := range files {
		// Skip directories and non-md files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Extract timestamp from filename (remove .md extension)
		timestamp := strings.TrimSuffix(file.Name(), ".md")

		// Only add valid timestamp files (14 digits: yyyymmddhhmmss)
		if len(timestamp) == 14 && IsNumeric(timestamp) {
			versions = append(versions, file.Name())
		}
	}

	// If we don't have more versions than the max, no need to delete any
	if len(versions) <= maxVersions {
		return
	}

	// Sort versions by timestamp (oldest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] < versions[j]
	})

	// Delete oldest versions until we're at the max
	versionsToDelete := len(versions) - maxVersions
	for i := 0; i < versionsToDelete; i++ {
		versionPath := filepath.Join(versionDir, versions[i])
		if err := os.Remove(versionPath); err != nil {
			log.Printf("Error deleting old version %s: %v", versionPath, err)
		} else {
			log.Printf("Deleted old version: %s", versionPath)
		}
	}
}
