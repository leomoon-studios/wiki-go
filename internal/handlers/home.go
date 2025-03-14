package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"wiki-go/internal/config"
	"wiki-go/internal/i18n"
	"wiki-go/internal/types"
	"wiki-go/internal/utils"
)

// Default homepage content
const defaultHomepageContent = `# Welcome to Wiki-Go

This is a simple yet powerful wiki platform built with Go. This page will guide you through everything you need to know to make the most of Wiki-Go.

## Important Configuration Notes

### Login Issues with Non-SSL Setups

If you're running Wiki-Go without SSL/HTTPS and experiencing login issues, you need to set ` + "`allow_insecure_cookies: true`" + ` in your ` + "`config.yaml`" + ` file. This is because:

1. By default, Wiki-Go sets the "Secure" flag on cookies for security
2. Browsers reject "Secure" cookies on non-HTTPS connections
3. This prevents login from working properly on HTTP-only setups

> **Security Note**: Only use this setting in development or in trusted internal networks. For public-facing wikis, always use HTTPS.

## Getting Started

### Creating Content

* **Root documents**: Click the "New" button in the toolbar and leave the "Document Path" field empty. Just provide a title and slug.
* **Categorized documents**: Click "New", then specify a category in the "Document Path" field (e.g., "tutorials" or "projects/web").
* **Edit existing pages**: Click the "Edit" button in the toolbar to modify any page.

### Content Organization

Wiki-Go organizes content hierarchically with each directory containing a document.md file that stores the content for that page.

## Markdown Guide

Wiki-Go uses Markdown for formatting content. Here are some examples:

### Headings
| Markdown                | Rendered Output          |
|-------------------------|--------------------------|
| \# Heading level 1      | <h1>Heading level 1</h1> |
| \## Heading level 2     | <h2>Heading level 2</h2> |
| \### Heading level 3    | <h3>Heading level 3</h3> |
| \#### Heading level 4   | <h4>Heading level 4</h4> |
| \##### Heading level 5  | <h5>Heading level 5</h5> |
| \###### Heading level 6 | <h6>Heading level 6</h6> |

### Paragraphs
To create paragraphs, use a blank line to separate one or more lines of text.

### Line Breaks
To create a line break or new line (<br\>), end a line with two or more spaces, and then type return.

### Emphasis
You can add emphasis by making text bold or italic.

#### Bold
To bold text, add two asterisks or underscores before and after a word or phrase. To bold the middle of a word for emphasis, add two asterisks without spaces around the letters.

| Markdown                   | Rendered Output        |
|----------------------------|------------------------|
| Example \*\*bold\*\* text. | Example **bold** text. |
| Example \_\_bold\_\_ text. | Example __bold__ text. |
| Example\*\*bold\*\*text    | Example**bold**text    |

#### Italic
To italicize text, add one asterisk or underscore before and after a word or phrase. To italicize the middle of a word for emphasis, add one asterisk without spaces around the letters.

| Markdown                     | Rendered Output            |
|------------------------------|----------------------------|
| Example \*italicized\* text. | Example *italicized* text. |
| Example \_italicized\_ text. | Example _italicized_ text. |
| Example\*italicized\*text    | Example*italicized*text    |

#### Bold and Italic
To emphasize text with bold and italics at the same time, add three asterisks or underscores before and after a word or phrase. To bold and italicize the middle of a word for emphasis, add three asterisks without spaces around the letters.

| Markdown                                      | Rendered Output                         |
|-----------------------------------------------|-----------------------------------------|
| This text is \*\*\*really important\*\*\*.    | This text is ***really important***.    |
| This text is \_\_\_really important\_\_\_.    | This text is ___really important___.    |
| This text is \_\_\*really important\*\_\_.    | This text is __*really important*__.    |
| This text is \*\*\_really important\_\*\*.    | This text is **_really important_**.    |
| This is really\*\*\*very\*\*\*important text. | This is really***very***important text. |

### Blockquotes
To create a blockquote, add a > in front of a paragraph.
` + "```text" + `
> Dorothy followed her through many of the beautiful rooms in her castle.
` + "```" + `
The rendered output looks like this:
> Dorothy followed her through many of the beautiful rooms in her castle.

#### Blockquotes with Multiple Paragraphs
Blockquotes can contain multiple paragraphs. Add a > on the blank lines between the paragraphs.
` + "``` text" + `
> Dorothy followed her through many of the beautiful rooms in her castle.
>
> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.
` + "```" + `
The rendered output looks like this:
> Dorothy followed her through many of the beautiful rooms in her castle.
>
> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.

#### Nested Blockquotes
Blockquotes can be nested. Add a >> in front of the paragraph you want to nest.
` + "```text" + `
> Dorothy followed her through many of the beautiful rooms in her castle.
>
>> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.
` + "```" + `
The rendered output looks like this:
> Dorothy followed her through many of the beautiful rooms in her castle.
>
>> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.

#### Blockquotes with Other Elements
Blockquotes can contain other Markdown formatted elements. Not all elements can be used — you'll need to experiment to see which ones work.
` + "```text" + `
> #### The quarterly results look great!
>
> - Revenue was off the chart.
> - Profits were higher than ever.
>
>  *Everything* is going according to **plan**.
` + "```" + `
The rendered output looks like this:
> #### The quarterly results look great!
>
> - Revenue was off the chart.
> - Profits were higher than ever.
>
>  *Everything* is going according to **plan**.

#### Lists and Task Lists

##### Regular Lists

Markdown supports both ordered and unordered lists. You can also nest lists to create sub-items.

**Unordered Lists** use asterisks (` + "`*`" + `), plus signs (` + "`+`" + `), or hyphens (` + "`-`" + `):

- Item 1
  - Sub-item 1
  - Sub-item 2
- Item 2
  - Sub-item 1
    - Sub-sub-item 1
    - Sub-sub-item 2
  - Sub-item 2
- Item 3

**Ordered Lists** use numbers followed by a period:

1. First item
   - Sub-item 1
   - Sub-item 2
2. Second item
   1. Sub-item 1
   2. Sub-item 2
3. Third item

You can also mix ordered and unordered lists:

1. First item
   - Sub-item 1
   - Sub-item 2
2. Second item
   - Sub-item 1
     1. Sub-sub-item 1
     2. Sub-sub-item 2
   - Sub-item 2
3. Third item

##### Task Lists

Markdown supports task lists, which are useful for tracking tasks or to-do items. Use square brackets to denote the state of each task: ` + "`[x]`" + ` for completed tasks and ` + "`[ ]`" + ` for incomplete tasks. Task lists can also include nested items.

**Example Task List:**

- [x] Write the press release
  - [x] Draft
  - [x] Review
  - [x] Finalize
- [ ] Update the website
  - [ ] Update home page
  - [ ] Update contact information

This task list shows that the press release has been fully completed, while the website update and media contact tasks are still pending with some sub-tasks.

### Extended Syntax
These are extended markdown features in Wiki-Go.

#### Text Highlight
To ==highlight text==, add two equal signs before and after a word or phrase. To highlight the middle of a word for emphasis, add two equal signs without spaces around the letters.

#### Superscript and Subscript
To create superscript text in Markdown, use the caret symbol (` + "`^`" + `). For example, ` + "`1^st^`" + ` renders as 1^st^. For subscript text, use the tilde symbol (` + "`~`" + `). For instance, ` + "`h~2~o`" + ` renders as h~2~o.

#### Strikethrough

To create strikethrough text in Markdown, use double tildes (` + "`~~`" + `). For example, ` + "`~~incorrect~~`" + ` renders as ~~incorrect~~.

#### Typographic Shortcodes
- ` + "`(c)`" + `: Replaced with © (Copyright symbol).
- ` + "`(r)`" + `: Replaced with ® (Registered trademark symbol).
- ` + "`(tm)`" + `: Replaced with ™ (Trademark symbol).
- ` + "`(p)`" + `: Replaced with ¶ (Paragraph symbol).
- ` + "`+-`" + `: Replaced with ± (Plus-minus symbol).
- ` + "`...`" + `: Replaced with … (Ellipsis).
- ` + "`<<`" + `: Replaced with « (Left angle quote).
- ` + "`>>`" + `: Replaced with » (Right angle quote).
- ` + "`1/2`" + `: Replaced with ½ (One-half).
- ` + "`1/4`" + `: Replaced with ¼ (One-quarter).
- ` + "`3/4`" + `: Replaced with ¾ (Three-quarters).

### Tables

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |

### Footnotes

Here's a sentence with a footnote.[^1]

[^1]: This is the footnote.

### Math Equations (MathJax)

Inline math: $E=mc^2$

Block math (requires blank lines before and after):

$$
\frac{d}{dx}(x^n) = nx^{n-1}
$$

### Diagrams (Mermaid)

` + "```mermaid" + `
graph TD;
    A-->B;
    A-->C;
    B-->D;
    C-->D;
` + "```" + `

## Advanced Features

### Details/Summary (Collapsible Content)

You can create collapsible sections using the details code fence:

` + "```details Details Title" + `
This is the collapsible content that will be hidden by default.

You can include any Markdown content here:
- Lists
- **Bold text**
- [Links](https://example.com)
- And more...
` + "```" + `

### Video Embedding

You can embed videos from various sources:

#### YouTube Videos:
` + "```youtube" + `
LcuvxJNIgfE
` + "```" + `

#### Vimeo Videos:
` + "```vimeo" + `
92060047
` + "```" + `

#### Local MP4 Files:

After uploading a video file through the attachments feature, you can insert it using files tab:
~~~
` + "```mp4" + `
your-video-filename.mp4
` + "```" + `
~~~

#### Forced RTL/LTR
You can force a specific direction for a section of text by adding the direction shortcode:
` + "```rtl" + `
متن نمونه...
` + "```" + `

### Shortcodes

Wiki-Go supports special shortcodes for dynamic content:

**Statistics Shortcode**:
` + "```markdown" + `
:::stats count=*:::
:::stats recent=5:::
` + "```" + `

These shortcodes display document statistics like total count or recent changes.

### File Attachments

1. Click "Edit" to enter edit mode
2. Click the "Attachments" button in the toolbar
3. Upload files and insert them into your document

### Version History

Wiki-Go keeps track of document changes. To view or restore previous versions:
1. Click "Edit" to enter edit mode
2. Click the "History" button in the toolbar

### Search

Use the search box in the sidebar to find content across your wiki.

### User Management

Administrators can manage users through Settings > Users.

### Language Selection

Change the interface language in Settings > Wiki.

### Theme Switching

Toggle between light and dark themes using the theme switch in the sidebar.

### Global Banner

Add a ` + "`banner.png`" + ` or ` + "`banner.jpg`" + ` file to your ` + "`data/static/`" + ` directory to display a global banner at the top of all pages.

## Features:

- **Fast**: Built with Go for high performance
- **Simple**: Minimal dependencies, easy to deploy
- **Embeddable**: All resources are embedded in the binary
- **Markdown Support**: Rich text formatting with extensions
- **Search**: Full-text search capabilities
- **Multi-user**: User management with admin privileges
- **Versioning**: Document version history
- **Responsive**: Mobile-friendly interface
- **Multilingual**: Interface translations
- **Dark Mode**: Light and dark theme support
- **File Attachments**: Upload and manage files

Need help? Check out the [GitHub repository](https://github.com/leomoon-studios/wiki-go) for more information.`

// EnsureHomepageExists creates the default homepage if it doesn't exist
func EnsureHomepageExists(cfg *config.Config) error {
	homepageDir := filepath.Join(cfg.Wiki.RootDir, "pages", "home")
	homepagePath := filepath.Join(homepageDir, "document.md")

	// Check if homepage directory exists, if not create it
	if _, err := os.Stat(homepageDir); os.IsNotExist(err) {
		if err := os.MkdirAll(homepageDir, 0755); err != nil {
			return fmt.Errorf("failed to create homepage directory: %w", err)
		}
	}

	// Check if homepage file exists, if not create it
	if _, err := os.Stat(homepagePath); os.IsNotExist(err) {
		if err := os.WriteFile(homepagePath, []byte(defaultHomepageContent), 0644); err != nil {
			return fmt.Errorf("failed to create homepage file: %w", err)
		}
		fmt.Println("Created default homepage at", homepagePath)
	}

	return nil
}

// HomeHandler renders the home page
func HomeHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// Add cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get navigation items
	nav, err := utils.BuildNavigation(cfg.Wiki.RootDir, cfg.Wiki.DocumentsDir)
	if err != nil {
		log.Printf("Error building navigation: %v", err)
		http.Error(w, "Failed to build navigation", http.StatusInternalServerError)
		return
	}

	// Mark active navigation item
	utils.MarkActiveNavItem(nav, "/")

	// Get the homepage path from the pages directory
	homepagePath := filepath.Join(cfg.Wiki.RootDir, "pages", "home", "document.md")

	// Always read the content from disk on each request to ensure
	// we display the most up-to-date version
	content, err := os.ReadFile(homepagePath)
	if err != nil {
		log.Printf("Error reading homepage: %v", err)
		// Fallback to a simple default if there's an error
		content = []byte("# Welcome to Wiki-Go\n\nThis is your homepage.")
	}

	// Get file information for last modified date
	docInfo, err := os.Stat(homepagePath)
	var lastModified time.Time
	if err == nil {
		lastModified = docInfo.ModTime()
	} else {
		lastModified = time.Now()
	}

	// Render the page
	data := &types.PageData{
		Navigation:         nav,
		Content:            template.HTML(utils.RenderMarkdown(string(content))),
		Breadcrumbs:        []types.BreadcrumbItem{{Title: "Home", Path: "/", IsLast: true}},
		Config:             cfg,
		LastModified:       lastModified,
		CurrentDir:         &types.NavItem{Title: "Home", Path: "/", IsDir: true, IsActive: true},
		AvailableLanguages: i18n.GetAvailableLanguages(),
	}

	renderTemplate(w, data)
}

// The renderTemplate and getTemplate functions have been moved to template.go
