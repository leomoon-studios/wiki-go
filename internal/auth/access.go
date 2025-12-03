package auth

import (
	"regexp"
	"strings"
	"wiki-go/internal/config"
)

// CanAccessDocument checks if the current session has access to the given document path
func CanAccessDocument(path string, session *Session, cfg *config.Config) bool {
	// Admin always has access
	if session != nil && session.Role == config.RoleAdmin {
		return true
	}

	// Find the first matching rule
	rule := findMatchingRule(path, cfg.AccessRules)

	// If no rule matches, default behavior depends on wiki privacy
	if rule == nil {
		if cfg.Wiki.Private {
			// If private, only authenticated users can access
			return session != nil
		}
		// If public, everyone can access
		return true
	}

	return checkAccessRule(rule, session)
}

func findMatchingRule(path string, rules []config.AccessRule) *config.AccessRule {
	for _, rule := range rules {
		if matchPattern(rule.Pattern, path) {
			return &rule
		}
	}
	return nil
}

func matchPattern(pattern, path string) bool {
	// Normalize path to ensure it starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	// Convert glob to regex
	var regexBuilder strings.Builder
	regexBuilder.WriteString("^")

	// We iterate through the pattern
	for i := 0; i < len(pattern); i++ {
		// Special handling for /** at the end of pattern
		// This allows /finance/** to match /finance, /finance/, and /finance/anything
		if i+3 <= len(pattern) && pattern[i:i+3] == "/**" && i+3 == len(pattern) {
			regexBuilder.WriteString("(/.*)?")
			i += 2 // Skip **
			break
		}

		if strings.HasPrefix(pattern[i:], "**") {
			regexBuilder.WriteString(".*")
			i++ // Skip the next *
		} else if pattern[i] == '*' {
			regexBuilder.WriteString("[^/]*")
		} else if pattern[i] == '?' {
			regexBuilder.WriteString("[^/]")
		} else {
			regexBuilder.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
	}
	regexBuilder.WriteString("$")

	matched, _ := regexp.MatchString(regexBuilder.String(), path)
	return matched
}

func checkAccessRule(rule *config.AccessRule, session *Session) bool {
	switch rule.Access {
	case "public":
		return true
	case "private":
		return session != nil
	case "restricted":
		if session == nil {
			return false
		}

		// Check groups
		for _, group := range rule.Groups {
			for _, userGroup := range session.Groups {
				if group == userGroup {
					return true
				}
			}
		}

		return false
	default:
		// Unknown access level, deny by default for safety
		return false
	}
}
