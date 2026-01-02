# Syntax

LeoMoon Wiki-Go uses Markdown for formatting content. Here are some examples:

## Headings
| Markdown                | Rendered Output          |
|-------------------------|--------------------------|
| \# Heading level 1      | <h1>Heading level 1</h1> |
| \## Heading level 2     | <h2>Heading level 2</h2> |
| \### Heading level 3    | <h3>Heading level 3</h3> |
| \#### Heading level 4   | <h4>Heading level 4</h4> |
| \##### Heading level 5  | <h5>Heading level 5</h5> |
| \###### Heading level 6 | <h6>Heading level 6</h6> |

## Paragraphs
To create paragraphs, use a blank line to separate one or more lines of text.

## Line Breaks
To create a line break or new line (<br\>), end a line with two or more spaces, and then type return.

## Emphasis
You can add emphasis by making text bold or italic.

### Bold
To bold text, add two asterisks or underscores before and after a word or phrase. To bold the middle of a word for emphasis, add two asterisks without spaces around the letters.

| Markdown                   | Rendered Output        |
|----------------------------|------------------------|
| Example \*\*bold\*\* text. | Example **bold** text. |
| Example \_\_bold\_\_ text. | Example __bold__ text. |
| Example\*\*bold\*\*text    | Example**bold**text    |

### Italic
To italicize text, add one asterisk or underscore before and after a word or phrase. To italicize the middle of a word for emphasis, add one asterisk without spaces around the letters.

| Markdown                     | Rendered Output            |
|------------------------------|----------------------------|
| Example \*italicized\* text. | Example *italicized* text. |
| Example \_italicized\_ text. | Example _italicized_ text. |
| Example\*italicized\*text    | Example*italicized*text    |

### Bold and Italic
To emphasize text with bold and italics at the same time, add three asterisks or underscores before and after a word or phrase. To bold and italicize the middle of a word for emphasis, add three asterisks without spaces around the letters.

| Markdown                                      | Rendered Output                         |
|-----------------------------------------------|-----------------------------------------|
| This text is \*\*\*really important\*\*\*.    | This text is ***really important***.    |
| This text is \_\_\_really important\_\_\_.    | This text is ___really important___.    |
| This text is \_\_\*really important\*\_\_.    | This text is __*really important*__.    |
| This text is \*\*\_really important\_\*\*.    | This text is **_really important_**.    |
| This is really\*\*\*very\*\*\*important text. | This is really***very***important text. |

## Blockquotes
To create a blockquote, add a > in front of a paragraph.
```text
> Dorothy followed her through many of the beautiful rooms in her castle.
```
The rendered output looks like this:
> Dorothy followed her through many of the beautiful rooms in her castle.

#### Blockquotes with Multiple Paragraphs
Blockquotes can contain multiple paragraphs. Add a > on the blank lines between the paragraphs.
``` text
> Dorothy followed her through many of the beautiful rooms in her castle.
>
> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.
```
The rendered output looks like this:
> Dorothy followed her through many of the beautiful rooms in her castle.
>
> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.

### Nested Blockquotes
Blockquotes can be nested. Add a >> in front of the paragraph you want to nest.
```text
> Dorothy followed her through many of the beautiful rooms in her castle.
>
>> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.
```
The rendered output looks like this:
> Dorothy followed her through many of the beautiful rooms in her castle.
>
>> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.

### Blockquotes with Other Elements
Blockquotes can contain other Markdown formatted elements. Not all elements can be used — you'll need to experiment to see which ones work.
```text
> #### The quarterly results look great!
>
> - Revenue was off the chart.
> - Profits were higher than ever.
>
>  *Everything* is going according to **plan**.
```
The rendered output looks like this:
> #### The quarterly results look great!
>
> - Revenue was off the chart.
> - Profits were higher than ever.
>
>  *Everything* is going according to **plan**.

### Lists and Task Lists

#### Regular Lists

Markdown supports both ordered and unordered lists. You can also nest lists to create sub-items.

**Unordered Lists** use asterisks (`*`), plus signs (`+`), or hyphens (`-`):

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

#### Task Lists

Markdown supports task lists, which are useful for tracking tasks or to-do items. Use square brackets to denote the state of each task: `[x]` for completed tasks and `[ ]` for incomplete tasks. Task lists can also include nested items.

**Example Task List:**

- [x] Write the press release
  - [x] Draft
  - [x] Review
  - [x] Finalize
- [ ] Update the website
  - [ ] Update home page
  - [ ] Update contact information

This task list shows that the press release has been fully completed, while the website update and media contact tasks are still pending with some sub-tasks.

Task lists also allows live editing if admin or editor user is logged in.

#### Kanban Boards

LeoMoon Wiki-Go supports interactive Kanban boards for project management. To create a kanban board, you need to:

1. **Set the document layout** in frontmatter:
   ```yaml
   ---
   layout: kanban
   ---
   ```

2. **Structure your content** using H4 headers for board titles (optional) and H5 headers for columns:

**Basic Kanban Structure:**
```markdown
---
layout: kanban
---

# Project Management Board

#### Sprint Planning

##### To Do
- [ ] Define project requirements
- [ ] Create user stories
- [ ] Set up development environment
  - [ ] Install dependencies
  - [ ] Configure database

##### In Progress
- [ ] Design user interface
- [ ] Implement authentication
  - [ ] User registration
  - [ ] Login functionality

##### Done
- [x] Project kickoff meeting
- [x] Team assignments
- [x] Initial planning
```

**Multiple Boards in One Document:**
```markdown
---
layout: kanban
---

# Development Project

#### Backend Development

##### To Do
- [ ] API design
- [ ] Database schema

##### In Progress
- [ ] User authentication

##### Done
- [x] Project setup

#### Frontend Development

##### To Do
- [ ] Component design
- [ ] Styling system

##### In Progress
- [ ] Login page

##### Done
- [x] Initial setup
```

**Key Features:**
- **Drag & Drop**: Tasks can be moved between columns by dragging
- **Live Editing**: Click on tasks to edit them inline
- **Nested Tasks**: Use indentation to create sub-tasks
- **Markdown Support**: Tasks support full markdown formatting
- **Duplicate Columns**: Column names can be duplicated without data loss
- **Multiple Boards**: Support for multiple kanban boards in one document

**Column Management:**
- H5 headers (`##### Column Name`) define kanban columns
- Columns are automatically converted to interactive drag-and-drop areas
- Empty columns can be deleted using the trash icon
- New columns can be added using the "+" button or by adding H5 headers in edit mode

**Task Management:**
- Tasks use standard markdown task list syntax: `- [ ]` for incomplete, `- [x]` for complete
- Tasks support all markdown formatting (bold, italic, links, code, etc.)
- Nested tasks are created using indentation (2 spaces per level)
- Tasks can be edited inline by clicking on them
- New tasks can be added using the "+" button in column headers

## Extended Syntax
These are extended markdown features in LeoMoon Wiki-Go.

### Text Highlight
To ==highlight text==, add two equal signs before and after a word or phrase. To highlight the middle of a word for emphasis, add two equal signs without spaces around the letters.

### Superscript and Subscript
To create superscript text in Markdown, use the caret symbol (`^`). For example, `1^st^` renders as 1^st^. For subscript text, use the tilde symbol (`~`). For instance, `h~2~o` renders as h~2~o.

### Strikethrough

To create strikethrough text in Markdown, use double tildes (`~~`). For example, `~~incorrect~~` renders as ~~incorrect~~.

### Typographic Shortcodes
- `(c)`: Replaced with (c) (Copyright symbol).
- `(r)`: Replaced with (r) (Registered trademark symbol).
- `(tm)`: Replaced with (tm) (Trademark symbol).
- `(p)`: Replaced with (p) (Paragraph symbol).
- `+-`: Replaced with +- (Plus-minus symbol).
- `...`: Replaced with ... (Ellipsis).
- `(1/2)`: Replaced with (1/2) (One-half).
- `(1/4)`: Replaced with (1/4) (One-quarter).
- `(3/4)`: Replaced with (3/4) (Three-quarters).

## Tables

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |

## Footnotes

Here's a sentence with a footnote.[^1]

[^1]: This is the footnote.

## Math Equations (MathJax)

Inline math: $E=mc^2$

Block math (requires blank lines before and after):

$$
\frac{d}{dx}(x^n) = nx^{n-1}
$$

## Diagrams (Mermaid)

```mermaid
graph TD;
    A-->B;
    A-->C;
    B-->D;
    C-->D;
```

## GitHub Flavored Alerts

You can use GitHub-style alerts to highlight important information.

> [!NOTE]
> Highlights information that users should take into account, even when skimming.

> [!TIP]
> Optional information to help a user be more successful.

> [!IMPORTANT]
> Crucial information necessary for users to succeed.

> [!WARNING]
> Critical content demanding immediate user attention due to potential risks.

> [!CAUTION]
> Negative potential consequences of an action.

## Details/Summary (Collapsible Content)

You can create collapsible sections using the details code fence:

```details Details Title
This is the collapsible content that will be hidden by default.

You can include any Markdown content here:
- Lists
- **Bold text**
- [Links](https://example.com)
- And more...
```

## Video Embedding

You can embed videos from various sources:

### YouTube Videos:
```youtube
LcuvxJNIgfE
```

### Vimeo Videos:
```vimeo
92060047
```

### Local MP4 Files:

After uploading a video file through the attachments feature, you can insert it using files tab:
~~~
```mp4
your-video-filename.mp4
```
~~~

### Forced RTL/LTR
You can force a specific direction for a section of text by adding the direction shortcode:
```rtl
Force RTL text.
```

## Shortcodes

LeoMoon Wiki-Go supports special shortcodes for dynamic content:

Use `[toc]` to automatically generate a table of contents based on the headings in your document.

Use `:::stats count=*:::` to display the total number of documents in your wiki, and `:::stats recent=5:::` to show a list of the 5 most recently modified documents.
