package pathidentity

import (
	"fmt"
	"runtime"
	"strings"
)

// ResolveStoragePath resolves an identity-derived internal storage path beneath
// an existing root, allowing missing destination components. Internal metadata
// must keep its assigned namespace: links cannot select another document's
// history or comments. Callers pass trusted names, never raw URL input.
func ResolveStoragePath(rootDir, logicalPath string) (string, error) {
	canonical, err := CanonicalDocumentRequestPath(logicalPath)
	if err != nil {
		return "", err
	}
	resolved, err := ResolveDocumentDestination(rootDir, ".", canonical)
	if err != nil {
		return "", err
	}
	same := canonical == resolved.LogicalPath
	if runtime.GOOS == "windows" {
		same = strings.EqualFold(canonical, resolved.LogicalPath)
	}
	if !same {
		return "", fmt.Errorf("internal storage identity changes its namespace")
	}
	return resolved.FilesystemPath, nil
}
