package goldext

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestDirectionFencesUseTypedNodesAndSafeMarkdown(t *testing.T) {
	markdown := newTrustedFenceTestMarkdown()
	source := []byte("```rtl\n# مرحبا\n\n**bold**\n\n<img src=x onerror=alert(1)>\n\n[bad](javascript:alert(2))\n```\n\n~~~ltr\nplain\n~~~\n")

	document := markdown.Parser().Parse(text.NewReader(source))
	var directions []TextDirection
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if direction, ok := node.(*DirectionBlock); ok {
				directions = append(directions, direction.Direction)
			}
		}
		return ast.WalkContinue, nil
	})
	if len(directions) != 2 || directions[0] != DirectionRTL || directions[1] != DirectionLTR {
		t.Fatalf("direction fences were not converted to fixed typed nodes: %#v", directions)
	}

	got := renderParsedDocument(t, markdown, source, document)
	for _, want := range []string{
		`<div class="rtl" dir="rtl">`,
		`<strong>bold</strong>`,
		`<div class="ltr" dir="ltr">`,
		`<p>plain</p>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("direction output omitted %q: %s", want, got)
		}
	}
	for _, forbidden := range []string{"<img", "onerror=", "javascript:"} {
		if strings.Contains(strings.ToLower(got), forbidden) {
			t.Errorf("direction output retained active content %q: %s", forbidden, got)
		}
	}
}

func TestInvalidDirectionFenceRemainsEscapedCode(t *testing.T) {
	markdown := newTrustedFenceTestMarkdown()
	source := []byte("```rtl onclick=alert(1)\n<img src=x onerror=alert(2)>\n```\n")
	var output bytes.Buffer
	if err := markdown.Convert(source, &output); err != nil {
		t.Fatalf("render invalid direction fence: %v", err)
	}

	got := output.String()
	if strings.Contains(got, `dir="rtl"`) || strings.Contains(got, `<div class="rtl"`) {
		t.Fatalf("invalid direction value created a direction node: %s", got)
	}
	if !strings.Contains(got, `<pre><code class="language-rtl">`) || !strings.Contains(got, `&lt;img src=x onerror=alert(2)&gt;`) {
		t.Fatalf("invalid direction fence was not rendered as escaped code: %s", got)
	}
}

func TestMermaidFenceEscapesSourceAndUsesStrictBrowserConfig(t *testing.T) {
	markdown := newTrustedFenceTestMarkdown()
	source := []byte("```mermaid\ngraph TD\nA-->B\nclick A \"javascript:alert(1)\"\n<img src=x onerror=alert(2)>\n```\n")
	document := markdown.Parser().Parse(text.NewReader(source))

	foundMermaid := false
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == KindMermaidBlock {
			foundMermaid = true
		}
		return ast.WalkContinue, nil
	})
	if !foundMermaid {
		t.Fatal("mermaid fence was not converted to a typed Mermaid node")
	}

	got := renderParsedDocument(t, markdown, source, document)
	if !strings.Contains(got, `<div class="mermaid">`) || !strings.Contains(got, `click A &quot;javascript:alert(1)&quot;`) {
		t.Fatalf("Mermaid source was not preserved as escaped text: %s", got)
	}
	if strings.Contains(strings.ToLower(got), "<img") || !strings.Contains(got, `&lt;img src=x onerror=alert(2)&gt;`) {
		t.Fatalf("Mermaid source produced active HTML: %s", got)
	}

	configPath := filepath.Join("..", "resources", "static", "js", "mermaid-init.js")
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read Mermaid browser configuration: %v", err)
	}
	initializeCount := bytes.Count(config, []byte("mermaid.initialize({"))
	strictCount := bytes.Count(config, []byte("securityLevel: 'strict'"))
	if initializeCount == 0 || strictCount != initializeCount {
		t.Fatalf("every Mermaid initialization must use strict security mode: initializations=%d strict=%d", initializeCount, strictCount)
	}
	if bytes.Contains(config, []byte("${err.message}")) || !bytes.Contains(config, []byte("errorElement.textContent")) {
		t.Fatal("Mermaid parser errors must be inserted as text, not interpolated HTML")
	}
}

func TestDirectionAndMermaidRendersDoNotShareState(t *testing.T) {
	const renderCount = 32
	markdown := newTrustedFenceTestMarkdown()
	results := make(chan string, renderCount)
	var waitGroup sync.WaitGroup

	for index := range renderCount {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			marker := fmt.Sprintf("render-marker-%d", index)
			source := fmt.Sprintf("```rtl\n**%s-direction**\n```\n\n```mermaid\ngraph TD\n%s-mermaid\n```\n", marker, marker)
			var output bytes.Buffer
			if err := markdown.Convert([]byte(source), &output, parser.WithContext(NewRenderContext(marker))); err != nil {
				results <- "error: " + err.Error()
				return
			}
			results <- output.String()
		}(index)
	}

	waitGroup.Wait()
	close(results)
	seen := make(map[string]bool, renderCount)
	for output := range results {
		if strings.HasPrefix(output, "error: ") {
			t.Fatal(output)
		}
		for index := range renderCount {
			marker := fmt.Sprintf("render-marker-%d", index)
			if strings.Contains(output, marker+"-direction") {
				if seen[marker] {
					t.Fatalf("render marker appeared in multiple outputs: %s", marker)
				}
				seen[marker] = true
				if !strings.Contains(output, marker+"-mermaid") {
					t.Fatalf("direction and Mermaid content became separated: %s", output)
				}
			}
		}
	}
	if len(seen) != renderCount {
		t.Fatalf("expected %d isolated renders, found %d", renderCount, len(seen))
	}
}

func newTrustedFenceTestMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(TrustedNodes),
	)
}

func renderParsedDocument(t *testing.T, markdown goldmark.Markdown, source []byte, document ast.Node) string {
	t.Helper()
	var output bytes.Buffer
	if err := markdown.Renderer().Render(&output, source, document); err != nil {
		t.Fatalf("render parsed document: %v", err)
	}
	return output.String()
}
