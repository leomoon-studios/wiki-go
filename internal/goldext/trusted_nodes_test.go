package goldext

import (
	"bytes"
	"fmt"
	"html"
	"strings"
	"sync"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
)

func TestTrustedNodesRenderFixedEscapedMarkup(t *testing.T) {
	document := ast.NewDocument()
	block := NewTrustedBlock(TrustedBlockDiv)
	block.Class = `safe" onclick="alert(1)`
	block.ID = `Overview" onmouseover="alert(2)`
	block.Direction = `RTL`
	block.AriaLabel = `Summary" autofocus onfocus="alert(4)`
	block.AppendChild(block, ast.NewString([]byte(`<img src=x onerror="alert(5)">`)))

	mark := NewTrustedInline(TrustedInlineMark)
	mark.Title = `Important" onmouseenter="alert(6)`
	mark.AppendChild(mark, ast.NewString([]byte(`<script>alert(7)</script>`)))
	block.AppendChild(block, mark)
	document.AppendChild(document, block)

	got := renderTrustedTree(t, document)
	lowerOutput := strings.ToLower(got)

	for _, forbidden := range []string{` onclick="`, ` onmouseover="`, ` onload="`, ` onfocus="`, ` onmouseenter="`, `<img`, `<script`} {
		if strings.Contains(lowerOutput, forbidden) {
			t.Errorf("trusted renderer emitted active content %q: %s", forbidden, got)
		}
	}
	for _, want := range []string{
		`<div class="safe onclick-alert-1"`,
		`id="overview-onmouseover-alert-2"`,
		`dir="rtl"`,
		`aria-label="Summary&#34; autofocus onfocus=&#34;alert(4)"`,
		`&lt;img src=x onerror=&quot;alert(5)&quot;&gt;`,
		`<mark title="Important&#34; onmouseenter=&#34;alert(6)">`,
		`&lt;script&gt;alert(7)&lt;/script&gt;`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("trusted renderer omitted %q: %s", want, got)
		}
	}
}

func TestTrustedAnchorValidatesDestinations(t *testing.T) {
	document := ast.NewDocument()
	block := NewTrustedBlock(TrustedBlockDiv)
	document.AppendChild(document, block)

	valid := NewTrustedInline(TrustedInlineAnchor)
	valid.Href = "https://example.com/docs?q=wiki"
	valid.AppendChild(valid, ast.NewString([]byte("valid")))
	block.AppendChild(block, valid)

	invalid := NewTrustedInline(TrustedInlineAnchor)
	invalid.Href = "javas&#x63;ript:alert(1)"
	invalid.AppendChild(invalid, ast.NewString([]byte("invalid")))
	block.AppendChild(block, invalid)

	got := renderTrustedTree(t, document)
	if !strings.Contains(got, `href="https://example.com/docs?q=wiki" rel="nofollow noreferrer"`) {
		t.Fatalf("trusted renderer omitted approved URL: %s", got)
	}
	if strings.Contains(strings.ToLower(html.UnescapeString(got)), "javascript:") {
		t.Fatalf("trusted renderer retained dangerous URL: %s", got)
	}
	if !strings.Contains(got, `<a>invalid</a>`) {
		t.Fatalf("trusted renderer did not fail closed for invalid URL: %s", got)
	}
}

func TestTrustedNodesCannotBeCreatedByUserMarkdown(t *testing.T) {
	markdown := goldmark.New(goldmark.WithExtensions(TrustedNodes))
	input := `<div class="trusted" onclick="alert(1)">raw</div>

<mark onmouseover="alert(2)">mark</mark>

WikiGoTrustedBlock`

	var output bytes.Buffer
	if err := markdown.Convert([]byte(input), &output); err != nil {
		t.Fatalf("render Markdown: %v", err)
	}

	got := strings.ToLower(html.UnescapeString(output.String()))
	for _, forbidden := range []string{"<div", "<mark", "onclick", "onmouseover"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("user Markdown created trusted markup %q: %s", forbidden, output.String())
		}
	}
	if !strings.Contains(got, "wikigotrustedblock") {
		t.Errorf("ordinary text imitating a node name was unexpectedly removed: %s", output.String())
	}
}

func TestInvalidTrustedNodeFailsClosed(t *testing.T) {
	document := ast.NewDocument()
	block := NewTrustedBlock(TrustedBlockInvalid)
	block.AppendChild(block, ast.NewString([]byte("must not render")))
	document.AppendChild(document, block)

	if got := renderTrustedTree(t, document); got != "" {
		t.Fatalf("invalid trusted node rendered output: %q", got)
	}
}

func renderTrustedTree(t *testing.T, document ast.Node) string {
	t.Helper()
	markdown := goldmark.New(goldmark.WithExtensions(TrustedNodes))
	var output bytes.Buffer
	if err := markdown.Renderer().Render(&output, nil, document); err != nil {
		t.Fatalf("render trusted tree: %v", err)
	}
	return output.String()
}

func TestTrustedValueHelpers(t *testing.T) {
	if got := EscapeHTMLText(`<tag title="x">&`); got != `&lt;tag title=&#34;x&#34;&gt;&amp;` {
		t.Errorf("EscapeHTMLText() = %q", got)
	}
	if got := NormalizeIdentifier(`  Board "Roadmap" / Q3  `); got != "board-roadmap-q3" {
		t.Errorf("NormalizeIdentifier() = %q", got)
	}
	if got := NormalizeClassList(" wiki-toc  alert<script> "); got != "wiki-toc alert-script" {
		t.Errorf("NormalizeClassList() = %q", got)
	}
}

func TestValidateTrustedURL(t *testing.T) {
	allowed := []string{
		"/api/files/guide/image.png",
		"../nearby",
		"guide/page?q=wiki#part",
		"#heading",
		"https://example.com/path",
		"http://example.com/path",
		"mailto:security@example.com",
	}
	for _, value := range allowed {
		t.Run("allow/"+value, func(t *testing.T) {
			if _, ok := ValidateTrustedURL(value); !ok {
				t.Errorf("ValidateTrustedURL(%q) rejected approved URL", value)
			}
		})
	}

	rejected := []string{
		"javascript:alert(1)",
		"JaVaScRiPt:alert(1)",
		"javas&#x63;ript:alert(1)",
		"javas&amp;#x63;ript:alert(1)",
		"data:text/html,x",
		"vbscript:msgbox(1)",
		"//evil.example/path",
		`\\evil.example\path`,
		"https://user:password@example.com/path",
		"https:example.com",
		"https://example.com/path\nnext",
	}
	for _, value := range rejected {
		t.Run("reject/"+value, func(t *testing.T) {
			if got, ok := ValidateTrustedURL(value); ok {
				t.Errorf("ValidateTrustedURL(%q) = (%q, true)", value, got)
			}
		})
	}
}

func TestRenderContextStateIsPerRenderAndConcurrentSafe(t *testing.T) {
	const renders = 32
	var waitGroup sync.WaitGroup
	results := make(chan string, renders)

	for index := range renders {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			context := NewRenderContext(fmt.Sprintf("document-%d", index))
			state := RenderStateFromContext(context)
			state.Set("owner", index)
			owner, ok := state.Get("owner")
			if !ok || owner != index {
				results <- "state leak"
				return
			}
			results <- state.DocumentPath + ":" + state.NextID("block") + ":" + state.NextID("block")
		}(index)
	}

	waitGroup.Wait()
	close(results)
	seen := make(map[string]bool, renders)
	for result := range results {
		if result == "state leak" {
			t.Fatal("render context leaked state between renders")
		}
		if seen[result] {
			t.Fatalf("duplicate render state result %q", result)
		}
		seen[result] = true
		if !strings.HasSuffix(result, ":block-1:block-2") {
			t.Errorf("render-local sequence did not start independently: %s", result)
		}
	}
}
