package goldext

import (
	"net/url"
	"regexp"
	"strings"
)

var vimeoIDPattern = regexp.MustCompile(`^[0-9]{1,20}$`)

// ExtractVimeoID accepts an exact numeric Vimeo ID or a URL on an approved
// Vimeo host. It returns an empty string for malformed or spoofed inputs.
func ExtractVimeoID(input string) string {
	input = strings.TrimSpace(input)
	if vimeoIDPattern.MatchString(input) {
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
	case "vimeo.com", "www.vimeo.com":
		candidate = singlePathValue(parsed.Path)
	case "player.vimeo.com":
		if strings.HasPrefix(parsed.Path, "/video/") {
			candidate = singlePathValue(strings.TrimPrefix(parsed.Path, "/video"))
		}
	default:
		return ""
	}

	if !vimeoIDPattern.MatchString(candidate) {
		return ""
	}
	return candidate
}
