# Features

## Features at a Glance

- ✍️ Full Markdown editing with emoji, Mermaid diagrams, and LaTeX math
- 🔍 Smart full-text search with highlighting and advanced filters
- 📁 Hierarchical page structure with version history
- 👥 User management, access control, and private wiki mode
- 🔗 Link management with automatic metadata fetching and categorization
- 💬 Comments with moderation and markdown support
- 📋 Interactive Kanban boards for project management
- ⚡ Instant setup via Docker or prebuilt binaries
- 🧩 Custom logos, banners, shortcodes, and more

> Perfect for internal documentation, personal knowledge bases, team wikis, or project management.

## Content Management
- **Markdown Support**: Write content using Markdown syntax for rich formatting
- **Emoji Shortcodes**: Use emoji shortcodes like `:smile:` in your Markdown content
- **File Attachments**: Upload and manage images and documents (supports jpg, jpeg, png, webp, gif, svg, txt, log, csv, sfd, zip, pdf, docx, xlsx, pptx, mp4)
- **Link Management**: Create and organize collections of links with automatic metadata fetching, descriptions, and categorization
- **Hierarchical Organization**: Organize content in nested directories
- **Version History**: Track changes with full revision history and restore previous versions
- **Document Management**: Create, edit, and delete documents with a user-friendly interface
- **Document Sorting and Naming**: Control the order of documents in the sidebar through slug names:
  - Documents are sorted alphabetically by their directory slug name
  - Document titles (displayed in the sidebar and heading) are taken from the first H1 heading in document.md
  - To manually sort documents, prefix slug names with numbers (e.g., `1-overview`, `2-installation`, `3-usage`)
  - The slug name can differ from the displayed title, allowing for organized structure while maintaining readable titles
  - Example: A directory named `1-getting-started` with document.md containing `# Getting Started Guide` will show as "Getting Started Guide" in the sidebar but be sorted first

## Collaboration & Feedback
- **Comments System**: Enable discussions on documents with a full-featured commenting system
- **Markdown in Comments**: Format comments using the same Markdown syntax as in documents
- **Comment Moderation**: Administrators can delete inappropriate comments
- **Disable Comments**: Option to disable comments system-wide through the wiki settings

## Search & Navigation
- **Full-Text Search**: Powerful search functionality with support for:
  - Exact phrase matching (using quotes)
  - Inclusion/exclusion of terms
  - Highlighted search results
- **Breadcrumb Navigation**: Clear path visualization for easy navigation
- **Sidebar Navigation**: Quick access to document hierarchy

## User Experience
- **Responsive Design**: Works on desktop and mobile devices
- **Dark/Light Theme**: Toggle between dark and light modes
- **Code Syntax Highlighting**: Support for multiple programming languages
- **Math Rendering**: LaTeX math formula support via MathJax
- **Diagrams**: Mermaid diagram integration for creating flowcharts, sequence diagrams, etc.

## Administration
- **User Management**: Create and manage users with different permission levels
- **Admin Panel**: Configure wiki settings through a web interface
- **Statistics**: Track document metrics and site usage

## Advanced Features
- **Custom Shortcodes**: Extend markdown with special shortcodes like `:::stats recent=5:::` for additional functionality
- **Media Embedding**: Embed images, videos, and other media in your documents
- **Print Friendly**: Optimized printing support for documentation
- **API Access**: RESTful API for programmatic access to wiki content

## Project Management
- **Interactive Kanban Boards**: Transform any document into a visual project management board
- **Drag & Drop Tasks**: Move tasks between columns with intuitive drag-and-drop functionality
- **Multiple Boards**: Support for multiple kanban boards within a single document
- **Task Management**: Create, edit, and organize tasks with full markdown formatting support
- **Duplicate Column Support**: Allows duplicate column names to prevent data loss during editing
- **Real-time Updates**: Changes are automatically saved and synchronized
- **Nested Tasks**: Support for sub-tasks and hierarchical task organization
- **Task Status Tracking**: Visual indicators for task completion and progress
