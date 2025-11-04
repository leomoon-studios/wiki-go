/**
 * Editor Preview Module
 * Handles preview functionality and markdown rendering
 */

let previewElement = null;

// Function to create preview panel
function createPreview(container) {
    const preview = document.createElement('div');
    preview.className = 'editor-preview';
    container.appendChild(preview);
    previewElement = preview;
    return preview;
}

// Function to toggle preview
async function togglePreview() {
    const editor = window.EditorCore.getEditor();
    if (!editor || !previewElement) return;

    const isPreviewActive = previewElement.classList.contains('editor-preview-active');
    const editorElement = document.querySelector('.CodeMirror');
    const toolbar = document.querySelector('.custom-toolbar');

    if (isPreviewActive) {
        // Switch back to edit mode
        previewElement.classList.remove('editor-preview-active');

        // CRITICAL: Clear the preview HTML to free memory and prevent slowdowns
        previewElement.innerHTML = '';

        // Show editor again
        if (editorElement) {
            editorElement.style.display = 'block';
        }

        // Re-enable all toolbar buttons except preview
        if (toolbar) {
            const buttons = toolbar.querySelectorAll('button:not(.preview-button)');
            buttons.forEach(button => {
                button.disabled = false;
                button.classList.remove('disabled');
            });
        }

        // Change preview button icon to indicate function
        const previewButton = toolbar.querySelector('.preview-button i');
        if (previewButton) {
            previewButton.className = 'fa fa-eye';
        }

        // Make sure editor gets focus when returning from preview mode
        // Using a small timeout to ensure DOM is updated first
        setTimeout(() => {
            if (editor) {
                editor.refresh();
                editor.focus();
            }
        }, 50);
    } else {
        // Show preview
        const content = editor.getValue();
        await updatePreview(content);

        // Hide editor
        if (editorElement) {
            editorElement.style.display = 'none';
        }

        // Show preview
        previewElement.classList.add('editor-preview-active');

        // Disable all toolbar buttons except preview
        if (toolbar) {
            const buttons = toolbar.querySelectorAll('button:not(.preview-button)');
            buttons.forEach(button => {
                button.disabled = true;
                button.classList.add('disabled');
            });
        }

        // Change preview button icon to indicate function
        const previewButton = toolbar.querySelector('.preview-button i');
        if (previewButton) {
            previewButton.className = 'fa fa-edit';
            previewButton.parentElement.title = 'Back to Edit Mode';
        }
    }
}

// Function to update preview content
async function updatePreview(content) {
    if (!previewElement) return;

    try {
        // Show loading indicator
        previewElement.innerHTML = '<div class="preview-loading">Loading preview...</div>';

        // Get current path for handling relative links correctly
        const isHomepage = window.location.pathname === '/';
        const path = isHomepage ? '/' : window.location.pathname;

        // Check for frontmatter to add special styling if needed
        const hasFrontmatter = content.startsWith('---\n');

        // Call the server-side renderer
        const response = await fetch(`/api/render-markdown?path=${encodeURIComponent(path)}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'text/plain',
            },
            body: content
        });

        if (!response.ok) {
            throw new Error('Failed to render markdown');
        }

        const html = await response.text();

        // If the content has kanban board, add a class to the preview element
        if (hasFrontmatter && html.includes('kanban-board')) {
            previewElement.classList.add('kanban-preview');
        } else {
            previewElement.classList.remove('kanban-preview');
        }

        previewElement.innerHTML = html;

        // Store Mermaid sources BEFORE any rendering happens
        const mermaidDiagrams = previewElement.querySelectorAll('.mermaid');
        mermaidDiagrams.forEach((diagram) => {
            // Extract the original source from the rendered content
            const textContent = diagram.textContent || diagram.innerText;
            if (textContent && textContent.trim()) {
                diagram.dataset.mermaidSource = textContent.trim();
            }
        });

        // Initialize any client-side renderers (Prism, MathJax, etc)
        if (window.Prism) {
            Prism.highlightAllUnder(previewElement);
        }

        if (window.MathJax) {
            MathJax.typeset([previewElement]);
        }

        if (window.mermaid) {
            mermaid.init(undefined, previewElement.querySelectorAll('.mermaid'));
        }

    } catch (error) {
        console.error('Preview error:', error);
        previewElement.innerHTML = '<p>Error rendering preview</p>';
    }
}

// Cleanup function
function cleanup() {
    if (previewElement) {
        previewElement.remove();
        previewElement = null;
    }
}

// Export the module
window.EditorPreview = {
    createPreview,
    togglePreview,
    updatePreview,
    cleanup,
    
    // Getters
    getPreviewElement: () => previewElement
};