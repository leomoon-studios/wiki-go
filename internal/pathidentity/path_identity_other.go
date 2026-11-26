//go:build !windows

package pathidentity

import "path/filepath"

func validateIdentityPath(string) error {
	return nil
}

func existingPathIdentity(filename string) (string, error) {
	// Resolve links without folding case or applying Windows filename rules.
	return filepath.EvalSymlinks(filename)
}
