package goldext

import (
	"net/url"
	"regexp"
	"strings"
)

var youtubeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// ExtractYouTubeID accepts an exact YouTube ID or a URL on an approved
// YouTube host. It returns an empty string for malformed or spoofed inputs.
func ExtractYouTubeID(input string) string {
	input = strings.TrimSpace(input)
	if youtubeIDPattern.MatchString(input) {
		return input
	}
	if input == "" || strings.Contains(input, "\\") {
		return ""
	}

	parsed, err := url.Parse(input)
	if err != nil || parsed.User != nil || parsed.Fragment != "" || parsed.Port() != "" ||
		(parsed.Scheme != "https" && parsed.Scheme != "http") {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	var candidate string
	switch host {
	case "youtu.be", "www.youtu.be":
		candidate = singlePathValue(parsed.Path)
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		switch {
		case parsed.Path == "/watch":
			candidate = parsed.Query().Get("v")
		case strings.HasPrefix(parsed.Path, "/embed/"):
			candidate = singlePathValue(strings.TrimPrefix(parsed.Path, "/embed"))
		case strings.HasPrefix(parsed.Path, "/v/"):
			candidate = singlePathValue(strings.TrimPrefix(parsed.Path, "/v"))
		}
	default:
		return ""
	}

	if !youtubeIDPattern.MatchString(candidate) {
		return ""
	}
	return candidate
}

func singlePathValue(value string) string {
	value = strings.Trim(value, "/")
	if value == "" || strings.Contains(value, "/") {
		return ""
	}
	return value
}
