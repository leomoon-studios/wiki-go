package goldext

import (
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// wikiLinkRe matches [[target]] or [[target|label]]. The optional leading "!"
// is captured so Obsidian-style embeds (![[...]]) can be left untouched.
var wikiLinkRe = regexp.MustCompile(`(!?)\[\[([^\]\n]+)\]\]`)

// slugCache holds the last-built slug→path index and the mtime of
// documentsRoot at the time it was built. A single os.Stat on documentsRoot
// per render is enough to detect any structural change (create/delete/rename).
var slugCache struct {
	mu      sync.RWMutex
	index   map[string]string
	modTime time.Time
}

// cachedSlugIndex returns the slug→path index, rebuilding it only when the
// mtime of documentsRoot has changed since the last build.
func cachedSlugIndex() map[string]string {
	documentsRoot := DocumentsRoot()
	info, err := os.Stat(documentsRoot)
	var mtime time.Time
	if err == nil {
		mtime = info.ModTime()
	}

	// Fast path: cache hit under read-lock.
	slugCache.mu.RLock()
	if slugCache.index != nil && slugCache.modTime.Equal(mtime) {
		idx := slugCache.index
		slugCache.mu.RUnlock()
		return idx
	}
	slugCache.mu.RUnlock()

	// Slow path: rebuild under write-lock with double-checked locking.
	slugCache.mu.Lock()
	defer slugCache.mu.Unlock()
	if slugCache.index != nil && slugCache.modTime.Equal(mtime) {
		return slugCache.index // another goroutine beat us here
	}
	slugCache.index = buildSlugIndex()
	slugCache.modTime = mtime
	return slugCache.index
}

// WikiLinkPreprocessor converts [[wikilinks]] into standard Markdown links
// before Goldmark runs. Examples:
//
//	[[02-entities/mass-spec]]            -> [mass-spec](/02-entities/mass-spec)
//	[[mass-spec]]                        -> [mass-spec](/02-entities/mass-spec)  (resolved by name)
//	[[02-entities/mass-spec|Mass Spec]]  -> [Mass Spec](/02-entities/mass-spec)
//	[[notes#results]]                    -> [notes](/notes#results)
//	[[#summary]]                         -> [summary](#summary)
//
// Targets that contain a slash are treated as explicit paths rooted at the wiki
// root. A bare target (no slash) is resolved by name against the whole document
// tree, wherever the page is filed (shortest path wins on duplicates), mirroring
// Obsidian-style linking; if no page by that name exists it falls back to a
// root-level path. Display text defaults to the last path segment. Content
// inside code spans/blocks and the ![[...]] embed form are left untouched.
//
// Note: docPath is accepted to satisfy the Preprocessor interface but is not
// currently used — bare-name resolution is always wiki-root-relative. It is a
// candidate for future document-relative resolution (Obsidian vault-relative).
func WikiLinkPreprocessor(markdown string, docPath string) string {
	if !strings.Contains(markdown, "[[") {
		return markdown
	}

	sections := splitCodeSections(markdown)

	// Resolve a bare name via the cached slug index (rebuilt only when
	// configured documents-root mtime changes).
	resolveName := func(name string) string {
		return cachedSlugIndex()[name]
	}

	for i := range sections {
		if sections[i].isCode {
			continue
		}
		content := sections[i].content
		matches := wikiLinkRe.FindAllStringSubmatchIndex(content, -1)
		if len(matches) == 0 {
			continue
		}
		var sb strings.Builder
		last := 0
		for _, m := range matches {
			// m[0]:m[1] = full match, m[2]:m[3] = group 1 (!), m[4]:m[5] = group 2 (target)
			sb.WriteString(content[last:m[0]])
			last = m[1]

			embedPrefix := content[m[2]:m[3]]
			inner := content[m[4]:m[5]]

			// Leave embeds (![[...]]) for a future preprocessor to handle.
			if embedPrefix == "!" {
				sb.WriteString(content[m[0]:m[1]])
				continue
			}

			// Split target and optional display label on the first "|".
			target, label := inner, ""
			if idx := strings.Index(inner, "|"); idx != -1 {
				target, label = inner[:idx], inner[idx+1:]
			}
			target = strings.TrimSpace(target)
			label = strings.TrimSpace(label)
			if target == "" {
				sb.WriteString(escapeMarkdownText(content[m[0]:m[1]]))
				continue
			}

			// Separate any #anchor from the page path.
			pagePath, anchor := target, ""
			if idx := strings.Index(target, "#"); idx != -1 {
				pagePath, anchor = target[:idx], target[idx:]
			}

			// Default display text is the last path segment (anchor stripped).
			if label == "" {
				label = pagePath
				if idx := strings.LastIndex(pagePath, "/"); idx != -1 {
					label = pagePath[idx+1:]
				}
				if label == "" {
					label = strings.TrimPrefix(anchor, "#")
				}
			}

			// Resolve a bare name (no slash, not an in-page anchor, not already
			// absolute) by looking it up anywhere in the document tree.
			if pagePath != "" && !strings.HasPrefix(pagePath, "/") && !strings.Contains(pagePath, "/") {
				if resolved := resolveName(pagePath); resolved != "" {
					pagePath = resolved
				}
			}

			urlValue, ok := safeWikiTarget(pagePath, anchor)
			if !ok {
				sb.WriteString(escapeMarkdownText(content[m[0]:m[1]]))
				continue
			}
			sb.WriteString("[" + escapeMarkdownText(label) + "](" + urlValue + ")")
		}
		sb.WriteString(content[last:])
		sections[i].content = sb.String()
	}

	return joinSections(sections)
}

func safeWikiTarget(pagePath, anchor string) (string, bool) {
	pagePath = strings.TrimSpace(pagePath)
	anchor = strings.TrimSpace(strings.TrimPrefix(anchor, "#"))
	if strings.ContainsAny(pagePath, `\\?:#<>`) {
		return "", false
	}
	for _, character := range pagePath {
		if character < ' ' || character == 0x7f {
			return "", false
		}
	}

	pagePath = strings.Trim(pagePath, "/")
	encodedSegments := make([]string, 0)
	if pagePath != "" {
		for _, segment := range strings.Split(pagePath, "/") {
			if segment == "" || segment == "." || segment == ".." {
				return "", false
			}
			encodedSegments = append(encodedSegments, url.PathEscape(segment))
		}
	}

	result := ""
	if len(encodedSegments) > 0 {
		result = "/" + strings.Join(encodedSegments, "/")
	}
	if anchor != "" {
		normalizedAnchor := NormalizeIdentifier(anchor)
		if normalizedAnchor == "" {
			return "", false
		}
		result += "#" + normalizedAnchor
	}
	return result, result != ""
}

func escapeMarkdownText(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		"[", `\[`,
		"]", `\]`,
		"(", `\(`,
		")", `\)`,
		"`", "\\`",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

// buildSlugIndex walks the documents tree and maps each page's slug (its
// directory name) to its path relative to the documents root. On duplicate
// slugs the shortest path wins (fewest segments, then lexically smallest),
// mirroring Obsidian's shortest-path name resolution.
func buildSlugIndex() map[string]string {
	index := make(map[string]string)
	documentsRoot := DocumentsRoot()
	filepath.WalkDir(documentsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "document.md" {
			return nil
		}
		rel, err := filepath.Rel(documentsRoot, filepath.Dir(p))
		if err != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		slug := path.Base(rel)
		if existing, ok := index[slug]; !ok || shorterPath(rel, existing) {
			index[slug] = rel
		}
		return nil
	})
	return index
}

// shorterPath reports whether a should win over b as the resolution for a slug:
// fewer path segments first, then lexically smallest for determinism.
func shorterPath(a, b string) bool {
	if na, nb := strings.Count(a, "/"), strings.Count(b, "/"); na != nb {
		return na < nb
	}
	return a < b
}
