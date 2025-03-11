package utils

import (
	"os"
)

// GetFileInfo returns file information for the given path
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
