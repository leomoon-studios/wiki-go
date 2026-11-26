package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func validateIdentityPath(value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("Windows path is not valid UTF-8")
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" || component == "." {
			continue
		}
		if strings.HasSuffix(component, ".") || strings.HasSuffix(component, " ") || !filepath.IsLocal(component) {
			return fmt.Errorf("invalid Windows path component %q", component)
		}
		for _, character := range component {
			if character < 32 || strings.ContainsRune(`:<>"|?*`, character) {
				return fmt.Errorf("unsupported Windows path character in %q", component)
			}
		}
	}
	return nil
}

func existingPathIdentity(filename string) (string, error) {
	// On Windows EvalSymlinks uses FindFirstFile for each resolved component,
	// returning the stored long name and case, including for an 8.3 input alias.
	// It also resolves symlinks and junctions. Propagate all lookup failures.
	return filepath.EvalSymlinks(filename)
}
