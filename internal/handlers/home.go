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
	"wiki-go/internal/auth"
)

// Default homepage content
const defaultHomepageContent = `# Welcome to LeoMoon Wiki-Go

LeoMoon Wiki-Go is a modern, feature-rich, databaseless flat-file wiki platform built with Go. It provides a clean, intuitive interface for creating and managing knowledge bases, documentation, and collaborative content without requiring any external database.

## Important Configuration Note with Non-SSL Setups

If you're running LeoMoon Wiki-Go without SSL/HTTPS and experiencing login issues, you need to set ` + "`allow_insecure_cookies: true`" + ` in your ` + "`config.yaml`" + ` file. This is because:

1. By default, LeoMoon Wiki-Go sets the "Secure" flag on cookies for security
2. Browsers reject "Secure" cookies on non-HTTPS connections
3. This prevents login from working properly on HTTP-only setups

> **Security Note**: Only use this setting in development or in trusted internal networks. For public-facing wikis, always use HTTPS.

## Features

### Content Management
- **Markdown Support**: Write content using Markdown syntax for rich formatting
- **Emoji Shortcodes**: Use emoji shortcodes like ` + "`:::smile:::`" + ` in your Markdown content
- **File Attachments**: Upload and manage images and documents (supports jpg, jpeg, png, gif, svg, txt, log, csv, zip, pdf, docx, xlsx, pptx, mp4)
- **Hierarchical Organization**: Organize content in nested directories
- **Version History**: Track changes with full revision history and restore previous versions
- **Document Management**: Create, edit, and delete documents with a user-friendly interface

### Collaboration & Feedback
- **Comments System**: Enable discussions on documents with a full-featured commenting system
- **Markdown in Comments**: Format comments using the same Markdown syntax as in documents
- **Comment Moderation**: Administrators can delete inappropriate comments
- **Disable Comments**: Option to disable comments system-wide through the wiki settings

### Search & Navigation
- **Full-Text Search**: Powerful search functionality with support for:
  - Exact phrase matching (using quotes)
  - Inclusion/exclusion of terms
  - Highlighted search results
- **Breadcrumb Navigation**: Clear path visualization for easy navigation
- **Sidebar Navigation**: Quick access to document hierarchy

### User Experience
- **Responsive Design**: Works on desktop and mobile devices
- **Dark/Light Theme**: Toggle between dark and light modes
- **Code Syntax Highlighting**: Support for multiple programming languages
- **Math Rendering**: LaTeX math formula support via MathJax
- **Diagrams**: Mermaid diagram integration for creating flowcharts, sequence diagrams, etc.

### Administration
- **User Management**: Create and manage users with different permission levels
- **Admin Panel**: Configure wiki settings through a web interface
- **Statistics**: Track document metrics and site usage

### Advanced Features
- **Custom Shortcodes**: Extend markdown with special shortcodes like ` + "`:::stats recent=5:::`" + ` for additional functionality
- **Media Embedding**: Embed images, videos, and other media in your documents
- **Print Friendly**: Optimized printing support for documentation
- **API Access**: RESTful API for programmatic access to wiki content

## Customization

### Custom Favicon

LeoMoon Wiki-Go comes with default favicons, but you can easily replace them with your own:

1. To use custom favicons, place your files in the ` + "`data/static/`" + ` directory with the following names:
   - ` + "`favicon.ico`" + ` - Standard favicon format (used by older browsers)
   - ` + "`favicon.png`" + ` - PNG format favicon
   - ` + "`favicon.svg`" + ` - SVG format favicon (recommended for best quality at all sizes)

2. The application will automatically detect and use your custom favicon files without requiring a restart.

SVG format is recommended for favicons as it scales well to different sizes while maintaining crisp quality.

### Custom Logo (Optional)

You can add a custom logo to display in the sidebar above your wiki title:

1. Create a logo file in one of the supported formats:
   - ` + "`logo.svg`" + ` - SVG format (recommended for best quality)
   - ` + "`logo.png`" + ` - PNG format (alternative option)

2. Place the logo file in the ` + "`data/static/`" + ` directory.

3. The logo will automatically appear in the sidebar above your wiki title.

**Notes:**
- The logo is displayed at 120×120 pixels, but will maintain its aspect ratio
- SVG format is recommended for the best appearance at all screen sizes
- No configuration changes or application restart needed
- If no logo file is present, only the wiki title will be displayed
- If both logo.svg and logo.png exist, logo.svg will be used

### Global Banner (Optional)

You can add a banner image that will appear at the top of all documents:

1. Create a banner image in one of the supported formats:
   - ` + "`banner.png`" + ` - PNG format (recommended for best quality)
   - ` + "`banner.jpg`" + ` - JPG format (alternative option)

2. Place the banner file in the ` + "`data/static/`" + ` directory.

3. The banner will automatically appear at the top of all document content.

**Notes:**
- The banner is displayed with responsive width and a maximum height of 250px
- The banner maintains its aspect ratio while fitting different screen sizes
- No configuration changes or application restart needed
- To remove the banner, simply delete the file from the ` + "`data/static/`" + ` directory
- If both banner.png and banner.jpg exist, banner.png will be used

## Security

- **Authentication**: User authentication with secure password hashing
- **Private Mode**: Optional private wiki mode requiring login
- **Admin Controls**: Separate admin privileges for content management

## Usage

### User Management

LeoMoon Wiki-Go includes a user management system with different permission levels:

- **Admin users**: Can create, edit, and delete content, manage users, and change settings
- **Regular users**: Can view content (when in private mode)

The default admin credentials are:
- Username: ` + "`admin`" + `
- Password: ` + "`admin`" + `

It's recommended to change these credentials immediately after first login.

### Creating Content

1. Log in with admin credentials
2. Use the "New" button to create a new document
3. Write content using Markdown syntax
4. Save your document

### Organizing Content

LeoMoon Wiki-Go allows you to organize content in a hierarchical structure:

1. Create directories to group related documents
2. Use the move/rename feature to reorganize content
3. Navigate through your content using the sidebar or breadcrumbs

### Attaching Files

You can attach files to any document:

1. Navigate to the document
2. Click the "Files" tab
3. Upload files using the upload button
4. Insert links to files in your document

### Using Comments

The commenting system allows users to provide feedback and engage in discussions:

1. Navigate to any document
2. Scroll to the comments section at the bottom
3. Authenticated users can add comments using Markdown syntax
4. Administrators can delete any comments
5. Comments can be disabled system-wide through the admin settings panel

## Markdown Guide

LeoMoon Wiki-Go uses Markdown for formatting content. Here are some examples:

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
These are extended markdown features in LeoMoon Wiki-Go.

#### Text Highlight
To ==highlight text==, add two equal signs before and after a word or phrase. To highlight the middle of a word for emphasis, add two equal signs without spaces around the letters.

#### Superscript and Subscript
To create superscript text in Markdown, use the caret symbol (` + "`^`" + `). For example, ` + "`1^st^`" + ` renders as 1^st^. For subscript text, use the tilde symbol (` + "`~`" + `). For instance, ` + "`h~2~o`" + ` renders as h~2~o.

#### Strikethrough

To create strikethrough text in Markdown, use double tildes (` + "`~~`" + `). For example, ` + "`~~incorrect~~`" + ` renders as ~~incorrect~~.

#### Typographic Shortcodes
- ` + "`(c)`" + `: Replaced with (c) (Copyright symbol).
- ` + "`(r)`" + `: Replaced with (r) (Registered trademark symbol).
- ` + "`(tm)`" + `: Replaced with (tm) (Trademark symbol).
- ` + "`(p)`" + `: Replaced with (p) (Paragraph symbol).
- ` + "`+-`" + `: Replaced with +- (Plus-minus symbol).
- ` + "`...`" + `: Replaced with ... (Ellipsis).
- ` + "`1/2`" + `: Replaced with 1/2 (One-half).
- ` + "`1/4`" + `: Replaced with 1/4 (One-quarter).
- ` + "`3/4`" + `: Replaced with 3/4 (Three-quarters).

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

## Additional Markdown Features

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
Force RTL text.
` + "```" + `

### Shortcodes

LeoMoon Wiki-Go supports special shortcodes for dynamic content:

**Statistics Shortcode**:
` + "```" + `
:::stats count=*:::

:::stats recent=5:::
` + "```" + `

These shortcodes display document statistics like total count or recent changes.
`

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
		content = []byte("# Welcome to LeoMoon Wiki-Go\n\nThis is your homepage.")
	}

	// Get file information for last modified date
	docInfo, err := os.Stat(homepagePath)
	var lastModified time.Time
	if err == nil {
		lastModified = docInfo.ModTime()
	} else {
		lastModified = time.Now()
	}

	// Get authentication status
	session := auth.GetSession(r)
	isAuthenticated := session != nil
	
	// Get user role
	userRole := ""
	if isAuthenticated && session != nil {
		userRole = session.Role
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
		IsAuthenticated:    isAuthenticated,
		UserRole:           userRole,
	}

	renderTemplate(w, data)
}
