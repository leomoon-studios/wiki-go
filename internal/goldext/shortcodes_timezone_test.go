package goldext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeRecentDoc stages a document.md file in docDir with the given
// title and mod time so renderRecentEditsFromDir picks it up.
func writeRecentDoc(docDir, title string, modTime time.Time) error {
	if err := os.MkdirAll(docDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(docDir, "document.md")
	if err := os.WriteFile(path, []byte("# "+title+"\n"), 0o644); err != nil {
		return err
	}
	return os.Chtimes(path, modTime, modTime)
}

func TestFormatModTime(t *testing.T) {
	// Pick a moment that is in two different local dates depending on tz
	// 2024-01-15 23:30 UTC is 2024-01-16 00:30 in Europe/Berlin (UTC+1)
	// and 2024-01-15 15:30 in America/Los_Angeles.
	moment := time.Date(2024, 1, 15, 23, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		tz   string
		want string
	}{
		{name: "UTC", tz: "UTC", want: "2024-01-15 23:30"},
		{name: "Berlin", tz: "Europe/Berlin", want: "2024-01-16 00:30"},
		{name: "Los_Angeles", tz: "America/Los_Angeles", want: "2024-01-15 15:30"},
		{name: "Empty falls back to UTC", tz: "", want: "2024-01-15 23:30"},
		{name: "Invalid falls back to UTC", tz: "Not/AZone", want: "2024-01-15 23:30"},
	}

	prev := wikiTimezone
	t.Cleanup(func() { SetWikiTimezone(prev) })

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetWikiTimezone(tc.tz)
			got := formatModTime(moment, "2006-01-02 15:04")
			if got != tc.want {
				t.Fatalf("formatModTime(%q) = %q, want %q", tc.tz, got, tc.want)
			}
		})
	}
}

func TestRenderRecentEditsUsesWikiTimezone(t *testing.T) {
	prev := wikiTimezone
	t.Cleanup(func() { SetWikiTimezone(prev) })

	// Stage a single recent edit at a UTC time we can verify.
	dir := t.TempDir()
	docDir := dir + "/welcome"
	if err := writeRecentDoc(docDir, "Welcome", time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("writeRecentDoc: %v", err)
	}

	SetWikiTimezone("Asia/Tokyo") // UTC+9
	var buf strings.Builder
	renderRecentEditsFromDir(&buf, dir, 5)
	if !strings.Contains(buf.String(), "2024-06-01 19:00") {
		t.Fatalf("expected JST-formatted time in output, got: %s", buf.String())
	}

	SetWikiTimezone("UTC")
	buf.Reset()
	renderRecentEditsFromDir(&buf, dir, 5)
	if !strings.Contains(buf.String(), "2024-06-01 10:00") {
		t.Fatalf("expected UTC-formatted time in output, got: %s", buf.String())
	}
}
