package goldext

import (
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// documentsRoot is the directory Wiki-Go stores documents under, relative to
// the working directory at runtime (matching LinkPreprocessor's convention).
const documentsRoot = "data/documents"

// wikiLinkRe matches [[target]] or [[target|label]]. The optional leading "!"
// is captured so Obsidian-style embeds (![[...]]) can be left untouched.
var wikiLinkRe = regexp.MustCompile(`(!?)\[\[([^\]\n]+)\]\]`)

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
// root-level path so the link still red-links via LinkPreprocessor. Display text
// defaults to the last path segment. Content inside code spans/blocks and the
// ![[...]] embed form are left untouched.
func WikiLinkPreprocessor(markdown string, docPath string) string {
	sections := splitCodeSections(markdown)

	// Build the name->path index lazily, and only if a bare target appears.
	var index map[string]string
	var indexBuilt bool
	resolveName := func(name string) string {
		if !indexBuilt {
			index = buildSlugIndex()
			indexBuilt = true
		}
		return index[name]
	}

	for i := range sections {
		if sections[i].isCode {
			continue
		}
		sections[i].content = wikiLinkRe.ReplaceAllStringFunc(sections[i].content, func(match string) string {
			parts := wikiLinkRe.FindStringSubmatch(match)
			if len(parts) < 3 {
				return match
			}

			// Leave embeds (![[...]]) for a future preprocessor to handle.
			if parts[1] == "!" {
				return match
			}

			// Split target and optional display label on the first "|".
			target, label := parts[2], ""
			if idx := strings.Index(target, "|"); idx != -1 {
				target, label = target[:idx], target[idx+1:]
			}
			target = strings.TrimSpace(target)
			label = strings.TrimSpace(label)
			if target == "" {
				return match
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

			// Build the URL. A bare "#anchor" stays an in-page link; everything
			// else resolves to an absolute wiki path.
			url := pagePath
			if url != "" && !strings.HasPrefix(url, "/") {
				url = "/" + url
			}
			url += anchor

			return "[" + label + "](" + url + ")"
		})
	}

	return joinSections(sections)
}

// buildSlugIndex walks the documents tree and maps each page's slug (its
// directory name) to its path relative to the documents root. On duplicate
// slugs the shortest path wins (fewest segments, then lexically smallest),
// mirroring Obsidian's shortest-path name resolution.
func buildSlugIndex() map[string]string {
	index := make(map[string]string)
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
