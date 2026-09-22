package comments

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Comment represents a single comment on a document
type Comment struct {
	ID            string        // {timestamp}_{username}.md
	Author        string        // Extracted from filename
	Timestamp     string        // Extracted from filename (format: YYYYMMDDhhmmss)
	TimestampUnix int64         // Unix timestamp for sorting and processing
	Content       string        // Raw markdown content
	RenderedHTML  template.HTML // Rendered HTML (not stored, generated on read)
	FormattedTime string        // Formatted timestamp for display
}

// AddComment creates a new comment for a document
func AddComment(documentPath, content, username string) error {
	// Generate timestamp in YYYYMMDDhhmmss format
	timestamp := time.Now().Format("20060102150405")

	// Sanitize username for filename safety
	safeUsername := sanitizeUsername(username)

	// Create comment filename with sanitized username
	filename := fmt.Sprintf("%s_%s.md", timestamp, safeUsername)

	// Ensure comment directory exists
	commentDir := filepath.Join("data/comments", documentPath)
	if err := os.MkdirAll(commentDir, 0755); err != nil {
		return fmt.Errorf("failed to create comment directory: %w", err)
	}

	// Write comment content to file
	return os.WriteFile(filepath.Join(commentDir, filename), []byte(content), 0644)
}

// GetComments retrieves all comments for a document
func GetComments(documentPath string) ([]Comment, error) {
	// Read comment directory
	commentDir := filepath.Join("data/comments", documentPath)
	files, err := os.ReadDir(commentDir)
	if os.IsNotExist(err) {
		return []Comment{}, nil // No comments yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read comment directory: %w", err)
	}

	// Process each comment file
	comments := []Comment{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			// Parse filename to get timestamp and username
			parts := strings.SplitN(strings.TrimSuffix(file.Name(), ".md"), "_", 2)
			if len(parts) != 2 {
				continue // Invalid filename format
			}

			timestamp := parts[0]

			// Validate timestamp format (should be 14 digits)
			if len(timestamp) != 14 || !isNumeric(timestamp) {
				continue // Invalid timestamp format
			}

			// Convert to Unix timestamp for sorting purposes
			timestampUnix := parseTimestampToUnix(timestamp)
			if timestampUnix == 0 {
				continue // Could not parse timestamp
			}

			// Read comment content
			content, err := os.ReadFile(filepath.Join(commentDir, file.Name()))
			if err != nil {
				continue // Can't read file
			}

			comments = append(comments, Comment{
				ID:            file.Name(),
				Author:        parts[1], // Username from filename
				Timestamp:     timestamp,
				TimestampUnix: timestampUnix,
				Content:       string(content),
			})
		}
	}

	// Sort comments by timestamp (oldest first)
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].TimestampUnix < comments[j].TimestampUnix
	})

	return comments, nil
}

// DeleteComment deletes a comment if the user is an admin
func DeleteComment(commentID string, documentPath string, userIsAdmin bool) error {
	if !userIsAdmin {
		return errors.New("only admins can delete comments")
	}

	// Validate the comment ID to ensure it's safe
	if !isValidCommentID(commentID) {
		return errors.New("invalid comment ID")
	}

	// Delete the comment file
	commentPath := filepath.Join("data/comments", documentPath, commentID)
	return os.Remove(commentPath)
}

// Helper function to validate comment ID format (timestamp_username.md)
func isValidCommentID(id string) bool {
	if id == "" || filepath.Base(id) != id || strings.ContainsAny(id, `/\\`) || hasCommentFilenameControlCharacter(id) {
		return false
	}
	if !strings.HasSuffix(id, ".md") {
		return false
	}

	// Split and validate parts
	parts := strings.SplitN(strings.TrimSuffix(id, ".md"), "_", 2)
	if len(parts) != 2 {
		return false
	}

	username := parts[1]
	if username == "" || username == "." || username == ".." || sanitizeUsername(username) != username {
		return false
	}

	// Validate timestamp is in YYYYMMDDhhmmss format (14 digits).
	timestamp := parts[0]
	return len(timestamp) == 14 && isNumeric(timestamp)
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// parseTimestampToUnix converts a YYYYMMDDhhmmss timestamp to Unix timestamp
func parseTimestampToUnix(timestamp string) int64 {
	if len(timestamp) != 14 {
		return 0
	}

	year := timestamp[0:4]
	month := timestamp[4:6]
	day := timestamp[6:8]
	hour := timestamp[8:10]
	minute := timestamp[10:12]
	second := timestamp[12:14]

	timeStr := fmt.Sprintf("%s-%s-%sT%s:%s:%s", year, month, day, hour, minute, second)
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return 0
	}

	return t.Unix()
}

// FormatCommentTime formats a YYYYMMDDhhmmss timestamp for display
func FormatCommentTime(timestamp string) string {
	if len(timestamp) != 14 {
		return "Unknown date"
	}

	year := timestamp[0:4]
	month := timestamp[4:6]
	day := timestamp[6:8]
	hour := timestamp[8:10]
	minute := timestamp[10:12]
	second := timestamp[12:14]

	timeStr := fmt.Sprintf("%s-%s-%sT%s:%s:%s", year, month, day, hour, minute, second)
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return "Unknown date"
	}

	return t.Format("Jan 2, 2006 at 15:04")
}

// AreCommentsAllowed checks if comments are allowed for a document
func AreCommentsAllowed(content string) bool {
	// Look for the no-comments marker
	return !strings.Contains(strings.ToLower(content), "<!-- no comments -->")
}

// sanitizeUsername makes a username safe for use in filenames
func sanitizeUsername(username string) string {
	// Replace potentially problematic characters
	username = strings.ReplaceAll(username, " ", "_")
	username = strings.ReplaceAll(username, "#", "")
	username = strings.ReplaceAll(username, "%", "")
	username = strings.ReplaceAll(username, "&", "")
	username = strings.ReplaceAll(username, "{", "")
	username = strings.ReplaceAll(username, "}", "")
	username = strings.ReplaceAll(username, "\\", "")
	username = strings.ReplaceAll(username, ":", "")
	username = strings.ReplaceAll(username, "<", "")
	username = strings.ReplaceAll(username, ">", "")
	username = strings.ReplaceAll(username, "*", "")
	username = strings.ReplaceAll(username, "?", "")
	username = strings.ReplaceAll(username, "|", "")
	username = strings.ReplaceAll(username, "\"", "")
	username = strings.ReplaceAll(username, "'", "")
	username = strings.ReplaceAll(username, ";", "")
	username = strings.ReplaceAll(username, "/", "_")
	return username
}
