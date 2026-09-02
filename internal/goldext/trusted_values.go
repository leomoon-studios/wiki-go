package goldext

import (
	"html"
	"net/url"
	"strings"
	"unicode"
)

// EscapeHTMLText escapes text for insertion into HTML text or quoted
// attribute contexts. Renderers should still use fixed attribute names.
func EscapeHTMLText(value string) string {
	return html.EscapeString(value)
}

// NormalizeIdentifier converts user-derived labels to a conservative value
// suitable for an HTML id or class token.
func NormalizeIdentifier(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))

	var normalized strings.Builder
	lastWasSeparator := false
	for _, r := range value {
		allowed := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' || r == ':'
		if allowed {
			normalized.WriteRune(r)
			lastWasSeparator = false
			continue
		}
		if normalized.Len() > 0 && !lastWasSeparator {
			normalized.WriteByte('-')
			lastWasSeparator = true
		}
	}

	return strings.Trim(normalized.String(), "-")
}

// NormalizeClassList normalizes each whitespace-separated class token and
// drops empty tokens.
func NormalizeClassList(value string) string {
	classes := strings.Fields(value)
	normalized := make([]string, 0, len(classes))
	for _, class := range classes {
		if class = NormalizeIdentifier(class); class != "" {
			normalized = append(normalized, class)
		}
	}
	return strings.Join(normalized, " ")
}

// ValidateTrustedURL accepts local URLs and explicit http, https, and mailto
// URLs. It rejects scheme-relative URLs, browser-normalized backslashes,
// control characters, credentials, and encoded scheme tricks.
func ValidateTrustedURL(value string) (string, bool) {
	value = strings.TrimSpace(value)
	for range 3 {
		decoded := html.UnescapeString(value)
		if decoded == value {
			break
		}
		value = decoded
	}

	if value == "" || strings.HasPrefix(value, "//") || strings.Contains(value, "\\") {
		return "", false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return "", false
		}
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host != "" && parsed.Scheme == "" || parsed.User != nil {
		return "", false
	}

	switch strings.ToLower(parsed.Scheme) {
	case "":
		if parsed.Host != "" {
			return "", false
		}
	case "http", "https":
		if parsed.Host == "" {
			return "", false
		}
	case "mailto":
		if parsed.Opaque == "" {
			return "", false
		}
	default:
		return "", false
	}

	return parsed.String(), true
}
