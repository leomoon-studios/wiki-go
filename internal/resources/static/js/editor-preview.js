/**
 * Editor Preview Module
 * Handles preview functionality and markdown rendering
 */

let previewElement = null;
let editorMode = 'edit';
let modeBeforePreview = 'edit';
let splitDebounceTimer = null;

const PREVIEW_UPDATE_DELAY_MS = 300;

function schedulePreviewUpdate(callback) {
    clearTimeout(splitDebounceTimer);
    splitDebounceTimer = setTimeout(callback, PREVIEW_UPDATE_DELAY_MS);
}

function clearPreviewUpdateTimer() {
    if (splitDebounceTimer) {
        clearTimeout(splitDebounceTimer);
        splitDebounceTimer = null;
    }
}

function setToolbarButtonsEnabled(toolbar, enabled, excludeSelector = null) {
    if (!toolbar) return;
    const selector = excludeSelector ? `button:not(${excludeSelector})` : 'button';
    toolbar.querySelectorAll(selector).forEach(button => {
        button.classList.toggle('disabled', !enabled);
    });
}

function createPreview(container) {
    const preview = document.createElement('div');
    preview.className = 'editor-preview';
    container.appendChild(preview);
    previewElement = preview;
    return preview;
}

async function togglePreview() {
    const editor = window.EditorCore.getEditor();
    if (!editor || !previewElement) return;

    if (editorMode === 'preview') {
        editorMode = modeBeforePreview;
    } else {
        modeBeforePreview = editorMode;
        editorMode = 'preview';
    }

    await applyMode(editor);
    updateToolbarIcons(document.querySelector('.custom-toolbar'));
}

async function toggleSplit() {
    const editor = window.EditorCore.getEditor();
    if (!editor || !previewElement) return;

    editorMode = (editorMode === 'split') ? 'edit' : 'split';

    await applyMode(editor);
    updateToolbarIcons(document.querySelector('.custom-toolbar'));
}

async function applyMode(editor) {
    const mode = editorMode;
    const editorArea = document.querySelector('.editor-area');
    const editorElement = document.querySelector('.CodeMirror');
    const toolbar = document.querySelector('.custom-toolbar');

    editorArea.classList.toggle('split-mode', mode === 'split');
    previewElement.classList.toggle('editor-preview-active', mode !== 'edit');
    previewElement.classList.toggle('editor-preview-full', mode === 'preview');

    if (editorElement) {
        editorElement.style.display = mode === 'preview' ? 'none' : 'block';
    }

    if (mode !== 'edit') {
        await updatePreview(editor.getValue());
    } else {
        previewElement.innerHTML = '';
    }

    // Attach / remove preview→editor scroll listener
    if (previewElement._syncScrollHandler) {
        previewElement.removeEventListener('scroll', previewElement._syncScrollHandler);
        previewElement._syncScrollHandler = null;
    }
    if (mode === 'split' && window.EditorScrollSync) {
        previewElement._syncScrollHandler = () => {
            const lineMap = window.EditorPreview._lineMap;
            if (lineMap) {
                window.EditorScrollSync.syncPreviewToEditor(previewElement, lineMap, editor);
            }
        };
        previewElement.addEventListener('scroll', previewElement._syncScrollHandler);
    }

    if (mode === 'preview') {
        setToolbarButtonsEnabled(toolbar, false, '#toggle-preview');
    } else {
        setToolbarButtonsEnabled(toolbar, true);
    }

    if (mode !== 'preview') {
        setTimeout(() => {
            editor?.refresh();
            editor?.focus();
        }, 50);
    }
}



// Update toolbar button icons based on mode
function updateToolbarIcons(toolbar) {
    if (!toolbar) return;

    const previewButton = toolbar.querySelector('.preview-button i');
    const splitButton = toolbar.querySelector('#toggle-split');

    if (previewButton) {
        if (editorMode === 'edit' || editorMode === 'split') {
            previewButton.className = 'fa fa-eye';
            previewButton.parentElement.title = 'Toggle Preview (Ctrl+Shift+P)';
        } else {
            previewButton.className = 'fa fa-edit';
            previewButton.parentElement.title = 'Back to Edit Mode';
        }
    }

    if (splitButton) {
        const icon = splitButton.querySelector('i');
        if (editorMode === 'split') {
            icon.className = 'fa fa-compress';
            splitButton.title = 'Exit Split View (Ctrl+Shift+S)';
            splitButton.classList.add('active');
        } else {
            icon.className = 'fa fa-columns';
            splitButton.title = 'Toggle Split View (Ctrl+Shift+S)';
            splitButton.classList.remove('active');
        }
    }
}

// Function to update preview content
async function updatePreview(content) {
    if (!previewElement) return;

    try {
        // Get current path for handling relative links correctly
        const isHomepage = wikiPathname() === '/';
        const path = isHomepage ? '/' : wikiPathname();

        // Check for frontmatter to add special styling if needed
        const hasFrontmatter = content.startsWith('---\n');

        // Call the server-side renderer (request source-line annotations for scroll sync)
        const response = await fetch(wikiURL(`/api/render-markdown?path=${encodeURIComponent(path)}&source_lines=1`), {
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

        // Patch the DOM incrementally so scroll position is preserved
        if (typeof morphdom !== 'undefined') {
            const wrapper = document.createElement('div');
            wrapper.className = previewElement.className;
            wrapper.innerHTML = html;
            morphdom(previewElement, wrapper, { childrenOnly: true });
        } else {
            previewElement.innerHTML = html;
        }

        // Store Mermaid sources BEFORE any rendering happens
        const mermaidDiagrams = previewElement.querySelectorAll('.mermaid');
        mermaidDiagrams.forEach((diagram) => {
            // Extract the original source from the rendered content
            const textContent = diagram.textContent || diagram.innerText;
            if (textContent && textContent.trim()) {
                diagram.dataset.mermaidSource = textContent.trim();
            }
        });

        // Use lazy loader to ensure libraries are loaded before rendering
        const promises = [];

        // Load and initialize Prism if there are code blocks
        if (previewElement.querySelector('pre code')) {
            if (window.LazyLoader) {
                promises.push(window.LazyLoader.forceLoad('prism').then(() => {
                    if (window.Prism) {
                        Prism.highlightAllUnder(previewElement);
                    }
                }));
            } else if (window.Prism) {
                Prism.highlightAllUnder(previewElement);
            }
        }

        // Load and initialize MathJax if there are math formulas
        if (previewElement.querySelector('.math, .katex, [class*="math"]') || 
            previewElement.textContent.includes('$')) {
            if (window.LazyLoader) {
                promises.push(window.LazyLoader.forceLoad('mathjax').then(() => {
                    if (window.MathJax) {
                        MathJax.typeset([previewElement]);
                    }
                }));
            } else if (window.MathJax) {
                MathJax.typeset([previewElement]);
            }
        }

        // Load and initialize Mermaid if there are diagrams
        if (previewElement.querySelector('.mermaid')) {
            if (window.LazyLoader) {
                promises.push(window.LazyLoader.forceLoad('mermaid').then(() => {
                    if (window.mermaid) {
                        mermaid.init(undefined, previewElement.querySelectorAll('.mermaid'));
                    }
                }));
            } else if (window.mermaid) {
                mermaid.init(undefined, previewElement.querySelectorAll('.mermaid'));
            }
        }

        // Wait for all libraries to load and render
        await Promise.all(promises);

        // Rebuild the scroll-sync line map now that the DOM is updated
        if (window.EditorScrollSync) {
            window.EditorPreview._lineMap = window.EditorScrollSync.buildLineMap(previewElement);
        }

    } catch (error) {
        console.error('Preview error:', error);
        previewElement.innerHTML = '<p>Error rendering preview</p>';
    }
}

function cleanup() {
    if (previewElement) {
        previewElement.remove();
        previewElement = null;
    }
    editorMode = 'edit';
    modeBeforePreview = 'edit';
    clearPreviewUpdateTimer();
}

window.EditorPreview = {
    createPreview,
    togglePreview,
    toggleSplit,
    updatePreview,
    cleanup,
    getPreviewElement: () => previewElement,
    getEditorMode: () => editorMode,
    setEditorMode: (mode) => { editorMode = mode; },
    setDebounceTimer: schedulePreviewUpdate,
    clearDebounceTimer: clearPreviewUpdateTimer,
};

