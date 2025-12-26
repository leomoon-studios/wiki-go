package migration

import (
	"bytes"
	"log"
	"os"
)

// FixBrokenConfig checks for and fixes known configuration file corruption issues.
// Specifically, it looks for the "%!s(MISSING)" string that was erroneously generated
// by an older version of the config template.
func FixBrokenConfig(configPath string) error {
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file does not exist yet – nothing to fix.
			return nil
		}
		return err
	}

	// Check for the specific corruption string
	corruption := []byte("%!s(MISSING)")
	if !bytes.Contains(data, corruption) {
		return nil
	}

	log.Println("Detected broken config file (missing format argument). Fixing...")

	// Replace the corruption with an empty string
	// The corruption usually appears after "access_rules:\n", so removing it leaves "access_rules:\n"
	// which is valid YAML (null value).
	fixedData := bytes.ReplaceAll(data, corruption, []byte(""))

	// Write the fixed data back to the file
	if err := os.WriteFile(configPath, fixedData, 0644); err != nil {
		return err
	}

	log.Println("Config file fixed successfully.")
	return nil
}
