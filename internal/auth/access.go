package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"wiki-go/internal/config"
	"wiki-go/internal/utils"
)

// CanAccessDocument checks an existing document's resolved identity. Invalid or
// missing paths are denied for every caller, including administrators.
func CanAccessDocument(path string, session *Session, cfg *config.Config) bool {
	_, allowed, err := ResolveDocumentAccess(path, session, cfg)
	return err == nil && allowed
}

// ResolveDocumentAccess returns the identity used for both authorization and
// subsequent page reads. Callers must read FilesystemPath, not the input spelling.
func ResolveDocumentAccess(path string, session *Session, cfg *config.Config) (utils.ResolvedDocumentPath, bool, error) {
	resolved, err := resolveAccessIdentity(path, cfg, false)
	if err != nil {
		return utils.ResolvedDocumentPath{}, false, err
	}
	return resolved, canAccessIdentity(resolved.LogicalPath, session, cfg), nil
}

// CanCreateDocument explicitly authorizes a creation destination, resolving its
// existing parent without treating failed lookups as unrestricted documents.
// Mutation handlers must additionally enforce their role and parent policies.
func CanCreateDocument(path string, session *Session, cfg *config.Config) bool {
	resolved, err := resolveAccessIdentity(path, cfg, true)
	return err == nil && canAccessIdentity(resolved.LogicalPath, session, cfg)
}

func resolveAccessIdentity(path string, cfg *config.Config, destination bool) (utils.ResolvedDocumentPath, error) {
	if cfg == nil {
		return utils.ResolvedDocumentPath{}, fmt.Errorf("document configuration is unavailable")
	}
	resolver := utils.ResolveExistingDocumentPath
	if destination {
		resolver = utils.ResolveDocumentDestination
	}
	// Validate before selecting the homepage root, so cleaning cannot discard
	// a Windows namespace prefix or an invalid component.
	canonical, err := utils.CanonicalDocumentRequestPath(path)
	if err != nil {
		return utils.ResolvedDocumentPath{}, err
	}
	if canonical == "/" {
		resolved, err := resolver(cfg.Wiki.RootDir, filepath.Join("pages", "home"), "/")
		if errors.Is(err, os.ErrNotExist) {
			// The homepage has built-in content before its storage directory
			// exists. Resolve that one deliberate destination from the wiki root;
			// dangling links and other failures still fail closed.
			resolved, err = utils.ResolveDocumentDestination(cfg.Wiki.RootDir, ".", "/pages/home")
			resolved.LogicalPath = "/"
		}
		return resolved, err
	}
	resolved, err := resolver(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir, path)
	if err == nil && resolved.LogicalPath == "/" {
		return utils.ResolvedDocumentPath{}, fmt.Errorf("document alias resolves to the homepage namespace")
	}
	return resolved, err
}

func canAccessIdentity(path string, session *Session, cfg *config.Config) bool {
	// Admin always has access
	if session != nil && session.Role == config.RoleAdmin {
		return true
	}

	// Find the first matching rule
	rule := findMatchingRule(path, cfg.AccessRules)

	if rule == nil {
		if cfg.Wiki.Private {
			return session != nil
		}
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
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	path = strings.ReplaceAll(path, "\\", "/")
	// Normalize path to ensure it starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	// Special handling for root recursive pattern (/**)
	// In this wiki architecture, the homepage (/) is physically located in data/pages/home,
	// while all other documents are in data/documents.
	// A recursive rule on root (/**) should logically apply to the homepage and its children.
	// Since other documents are NOT children of the homepage in the file system,
	// we restrict /** to only match the root path itself.
	// This prevents a "Recursive Homepage" rule from matching the entire wiki.
	if pattern == "/**" {
		return path == "/"
	}

	// Convert glob to regex
	var regexBuilder strings.Builder
	if runtime.GOOS == "windows" {
		regexBuilder.WriteString("(?i)")
	}
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
