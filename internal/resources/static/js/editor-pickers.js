/**
 * Editor Pickers Module
 * Handles emoji, document, table, and anchor pickers
 */

// Picker element references
let tablePickerElement = null;
let emojiPickerElement = null;
let docPickerElement = null;
let anchorPickerElement = null;

// Data caches
let emojiData = [];
let documentData = [];

// Global emoji cache
const EmojiCache = {
    data: null,
    preloadPromise: null,

    // Get emoji data - returns a promise
    getData: function() {
        // If we already have data, return it
        if (this.data) {
            return Promise.resolve(this.data);
        }

        // If we're already loading, return the existing promise
        if (this.preloadPromise) {
            return this.preloadPromise;
        }

        // Start a new load
        this.preloadPromise = this.fetchEmojiData()
            .then(data => {
                this.data = data;
                return data;
            })
            .catch(err => {
                console.error('Error loading emoji data:', err);
                // Clear the promise so we can try again
                this.preloadPromise = null;
                return [];
            });

        return this.preloadPromise;
    },

    // Fetch emoji data from server
    fetchEmojiData: async function() {
        try {
            console.log('Fetching emoji data from server');
            const response = await fetch(wikiURL('/api/data/emojis'));
            if (!response.ok) {
                throw new Error('Failed to fetch emoji data');
            }
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Error fetching emoji data:', error);
            return [];
        }
    }
};

// ===== EMOJI PICKER =====

// Function to insert emoji
function insertEmoji(emoji) {
    const editor = window.EditorCore.getEditor();
    if (!editor) return;

    // Insert emoji shortcode at cursor position
    const cursor = editor.getCursor();
    editor.replaceRange(emoji.shortcode + ' ', cursor);

    // Focus the editor
    editor.focus();
}

// Create emoji picker
function createEmojiPicker() {
    const picker = document.createElement('div');
    picker.className = 'emoji-picker';
    picker.style.display = 'none';

    // Create a content container for the emojis
    const emojiContainer = document.createElement('div');
    emojiContainer.className = 'emoji-container';

    // Create a loading message
    const loadingMsg = document.createElement('div');
    loadingMsg.className = 'emoji-loading';
    loadingMsg.textContent = 'Loading emojis...';

    // Add both to the picker
    picker.appendChild(loadingMsg);
    picker.appendChild(emojiContainer);
    document.body.appendChild(picker);
    emojiPickerElement = picker;

    // Load emoji data using our singleton cache
    EmojiCache.getData().then(data => {
        // Save the data globally
        emojiData = data;

        // Show the emoji buttons and hide the loading message
        loadingMsg.style.display = 'none';

        // Create buttons for each emoji
        emojiData.forEach(emoji => {
            // Skip if no shortcodes
            if (!emoji.shortcodes || emoji.shortcodes.length === 0) return;

            const button = document.createElement('button');
            button.className = 'emoji-btn';

            // Get the primary shortcode (first in the array)
            const shortcode = ':' + emoji.shortcodes[0] + ':';
            button.title = shortcode;
            button.textContent = emoji.emoji;

            button.addEventListener('click', () => {
                const emojiObj = {
                    emoji: emoji.emoji,
                    shortcode: shortcode
                };
                insertEmoji(emojiObj);
                hideEmojiPicker();
            });
            emojiContainer.appendChild(button);
        });
    });

    return picker;
}

// Function to show emoji picker
function showEmojiPicker(button) {
    // Preload emoji data to reduce latency
    EmojiCache.getData();

    togglePicker({
        get: () => emojiPickerElement,
        set: (el) => { emojiPickerElement = el; },
        create: createEmojiPicker,
        button,
        closeSelector: '.emoji-button'
    });
}

// Function to hide emoji picker
function hideEmojiPicker() {
    if (emojiPickerElement) {
        emojiPickerElement.style.display = 'none';
        // Reset any transform or size constraints that might have been applied
        emojiPickerElement.style.transform = '';
    }
}

// ===== DOCUMENT PICKER =====

// Function to fetch documents
async function fetchDocuments() {
    try {
        const response = await fetch(wikiURL('/api/documents/list'));
        if (!response.ok) {
            if (response.status === 401) {
                throw new Error('Authentication required');
            }
            throw new Error('Failed to fetch documents');
        }
        const data = await response.json();
        if (!data.success) {
            throw new Error(data.message || 'Failed to fetch documents');
        }
        return data.documents || [];
    } catch (error) {
        console.error('Error fetching documents:', error);
        throw error; // Re-throw to allow handling in the caller
    }
}

// Function to insert document link
function insertDocLink(document) {
    const editor = window.EditorCore.getEditor();
    if (!editor) return;

    // Get selection (if any) to use as link text
    const selection = editor.getSelection();

    // Build path and fallback title
    let path = document.path;
    const title = document.title || path.split('/').pop();

    // Strip /documents prefix for internal links (except pages)
    if (path.startsWith('/documents')) {
        path = path.replace('/documents', '');
    }

    if (selection) {
        // Replace selection with link
        editor.replaceSelection(`[${selection}](${path})`);
        // Place cursor at end of link
    } else {
        const cursor = editor.getCursor();
        editor.replaceRange(`[${title}](${path})`, cursor);
        // Select the link text so user can change it
        editor.setSelection(
            { line: cursor.line, ch: cursor.ch + 1 },
            { line: cursor.line, ch: cursor.ch + 1 + title.length }
        );
    }

    editor.focus();
}

// Create document picker
function createDocPicker() {
    const picker = document.createElement('div');
    picker.className = 'doc-picker';
    picker.setAttribute('dir', 'auto');
    picker.style.display = 'none';

    // Create a search input
    const searchContainer = document.createElement('div');
    searchContainer.className = 'doc-search-container';

    const searchInput = document.createElement('input');
    searchInput.type = 'text';
    searchInput.className = 'doc-search-input';
    searchInput.placeholder = window.i18n ? window.i18n.t('docpicker.search_placeholder') : 'Search…';
    searchInput.style.boxSizing = 'border-box';
    searchInput.style.width = '100%';

    searchContainer.appendChild(searchInput);
    picker.appendChild(searchContainer);

    // Create a content container for the documents
    const docsContainer = document.createElement('div');
    docsContainer.className = 'docs-container';
    docsContainer.style.width = '100%';
    docsContainer.style.boxSizing = 'border-box';

    // Create a loading message
    const loadingMsg = document.createElement('div');
    loadingMsg.className = 'docs-loading';
    loadingMsg.textContent = window.i18n ? window.i18n.t('docpicker.loading') : 'Loading...';

    // Add both to the picker
    picker.appendChild(loadingMsg);
    picker.appendChild(docsContainer);
    document.body.appendChild(picker);
    docPickerElement = picker;

    // Check if user has editor or admin role before trying to fetch documents
    window.Auth.checkUserRole('editor').then(canEdit => {
        if (!canEdit) {
            loadingMsg.textContent = window.i18n ? window.i18n.t('docpicker.admin_required') : 'Editor or admin required';
            return;
        }

        // Load document data
        fetchDocuments().then(documents => {
            // Save the data globally
            documentData = documents;

            // Hide loading message
            loadingMsg.style.display = 'none';

            // Function to filter and display documents
            const filterAndDisplay = (query = '') => {
                // Clear previous documents
                docsContainer.innerHTML = '';

                // Filter documents based on search query
                const filteredDocs = query
                    ? documentData.filter(doc =>
                        includesIgnoreCase(doc.title, query) ||
                        includesIgnoreCase(doc.path, query))
                    : documentData;

                if (filteredDocs.length === 0) {
                    const noResults = document.createElement('div');
                    noResults.className = 'no-results';
                    noResults.textContent = window.i18n ? window.i18n.t('docpicker.no_results') : 'No results found';
                    docsContainer.appendChild(noResults);
                    return;
                }

                // Create buttons for each document
                filteredDocs.forEach(doc => {
                    const btn = createDocButton(doc, () => {
                        insertDocLink(doc);
                        hideDocPicker();
                    });
                    docsContainer.appendChild(btn);
                });
            };

            // Initial display of all documents
            filterAndDisplay();

            // Add search functionality
            searchInput.addEventListener('input', (e) => {
                filterAndDisplay(e.target.value);
            });
        }).catch(error => {
            // Show error message
            loadingMsg.textContent = window.i18n ? window.i18n.t('docpicker.error_loading').replace('{0}', error.message || 'Access denied') : 'Error loading documents';
        });
    });

    return picker;
}

// Function to show document picker
function showDocPicker(button) {
    togglePicker({
        get: () => docPickerElement,
        set: (el) => { docPickerElement = el; },
        create: createDocPicker,
        button,
        closeSelector: '.doc-link-button',
        onShow: (picker) => {
            const searchInput = picker.querySelector('.doc-search-input');
            if (searchInput) {
                setTimeout(() => searchInput.focus(), 50);
            }
        }
    });
}

// Function to hide document picker
function hideDocPicker() {
    if (docPickerElement) {
        docPickerElement.style.display = 'none';
        // Reset any transform or size constraints that might have been applied
        docPickerElement.style.transform = '';
    }
}

// ===== TABLE PICKER =====

// Table picker functionality
function createTablePicker() {
    const picker = document.createElement('div');
    picker.className = 'table-picker';
    picker.style.display = 'none';

    for (let i = 0; i < 10; i++) {
        for (let j = 0; j < 10; j++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.row = i + 1;
            cell.dataset.col = j + 1;
            cell.addEventListener('mouseover', () => highlightCells(cell));
            cell.addEventListener('click', () => insertTable(i + 1, j + 1));
            picker.appendChild(cell);
        }
        picker.appendChild(document.createElement('br'));
    }

    document.body.appendChild(picker);
    tablePickerElement = picker;
    return picker;
}

function highlightCells(cell) {
    const row = parseInt(cell.dataset.row);
    const col = parseInt(cell.dataset.col);
    const cells = document.querySelectorAll('.table-picker .cell');
    cells.forEach(c => {
        if (parseInt(c.dataset.row) <= row && parseInt(c.dataset.col) <= col) {
            c.classList.add('active');
        } else {
            c.classList.remove('active');
        }
    });
}

function insertTable(rows, cols) {
    const editor = window.EditorCore.getEditor();
    if (!editor) return;

    // Use the createMarkdownTable function from markdown-table-editor.js
    // which will create a properly formatted table
    const table = '\n' + (window.createMarkdownTable ? window.createMarkdownTable(rows, cols) : '');

    editor.replaceSelection(table);
    if (tablePickerElement) {
        tablePickerElement.style.display = 'none';
    }

    // Make sure editor gets focus after inserting table
    setTimeout(() => editor.focus(), 10);
}

function showTablePicker(button) {
    togglePicker({
        get: () => tablePickerElement,
        set: (el) => { tablePickerElement = el; },
        create: createTablePicker,
        button,
        closeSelector: '.table-button'
    });
}

// ===== ANCHOR PICKER =====

// Function to slugify heading text (mirror of Go makeSlug)
function makeSlug(text) {
    return text.toLowerCase()
        .replace(/[&+_,.()\[\]{}'"!?;:~*]/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
        .replace(/ /g, '-')
        .replace(/[^a-z0-9-]/g, '')
        .replace(/-+/g, '-')
        .replace(/^[-]+|[-]+$/g, '') || 'heading';
}

// Extract headings and IDs from current editor content
function extractAnchors(markdown) {
    const lines = markdown.split('\n');
    const headingRx = /^(#{1,6})\s+(.*?)\s*(\{#([a-zA-Z0-9\-]+)\})?\s*$/;
    const anchors = [];
    const usedIds = {};

    let inCodeBlock = false;

    lines.forEach((ln) => {
        const trimmed = ln.trim();

        // Toggle code block state on fences ``` or ~~~
        if (trimmed.startsWith('```') || trimmed.startsWith('~~~')) {
            inCodeBlock = !inCodeBlock;
            return;
        }
        if (inCodeBlock) return; // Skip lines inside fenced code blocks

        const m = trimmed.match(headingRx);
        if (!m) return;
        const level = m[1].length;
        const text = m[2].trim();
        let id = m[4] || makeSlug(text);
        if (!id) id = "heading";

        // Ensure uniqueness (mirror Go logic)
        if (usedIds[id]) {
            let counter = 1;
            let newId;
            do {
                newId = `${id}-${counter}`;
                counter++;
            } while (usedIds[newId]);
            id = newId;
        }
        usedIds[id] = true;

        anchors.push({ level, text, id });
    });

    return anchors;
}

// Insert anchor link in markdown
function insertAnchorLink(anchor) {
    const editor = window.EditorCore.getEditor();
    if (!editor) return;

    const selection = editor.getSelection();
    const linkText = selection || anchor.text;

    if (selection) {
        editor.replaceSelection(`[${selection}](#${anchor.id})`);
    } else {
        const cursor = editor.getCursor();
        editor.replaceRange(`[${linkText}](#${anchor.id})`, cursor);
        // Select link text for editing if no selection existed
        editor.setSelection(
            { line: cursor.line, ch: cursor.ch + 1 },
            { line: cursor.line, ch: cursor.ch + 1 + linkText.length }
        );
    }

    editor.focus();
}

// Helper: fetch raw markdown for a document path
async function fetchDocumentSource(docPath) {
    try {
        const response = await fetch(wikiURL(`/api/source${docPath}`));
        if (!response.ok) throw new Error('Failed to fetch document');
        return await response.text();
    } catch (err) {
        console.error('fetchDocumentSource', err);
        throw err;
    }
}

// Insert link to an anchor that belongs to another document
function insertDocAnchorLink(document, anchor) {
    const editor = window.EditorCore.getEditor();
    if (!editor) return;

    // Prepare path (mirror insertDocLink logic)
    let path = document.path;
    if (path.startsWith('/documents')) {
        path = path.replace('/documents', '');
    }

    const selection = editor.getSelection();
    const linkText = selection || anchor.text;
    const markdown = `[${linkText}](${path}#${anchor.id})`;

    if (selection) {
        editor.replaceSelection(markdown);
    } else {
        const cursor = editor.getCursor();
        editor.replaceRange(markdown, cursor);
        editor.setSelection(
            { line: cursor.line, ch: cursor.ch + 1 },
            { line: cursor.line, ch: cursor.ch + 1 + linkText.length }
        );
    }

    editor.focus();
}

// Create a picker that first lists documents and then their anchors
function createDocAnchorPicker() {
    const picker = document.createElement('div');
    picker.className = 'anchor-picker';
    picker.style.display = 'none';
    picker.setAttribute('dir', 'auto');

    // Header (title + back button)
    const header = document.createElement('div');
    header.className = 'anchor-picker-header';
    picker.appendChild(header);

    const backBtn = document.createElement('button');
    backBtn.className = 'anchor-back-btn';
    backBtn.textContent = '←';
    backBtn.style.display = 'none';
    header.appendChild(backBtn);

    const titleEl = document.createElement('span');
    titleEl.className = 'anchor-picker-title';
    titleEl.textContent = '';
    header.appendChild(titleEl);

    // Search input
    const searchInput = document.createElement('input');
    searchInput.type = 'text';
    searchInput.className = 'anchor-search-input';
    searchInput.placeholder = window.i18n ? window.i18n.t('docpicker.search_placeholder') : 'Search…';
    picker.appendChild(searchInput);

    // Container for list items
    const listContainer = document.createElement('div');
    listContainer.className = 'anchor-list-container';
    picker.appendChild(listContainer);

    document.body.appendChild(picker);

    // State variables
    let currentDocument = null;
    const anchorsCache = {}; // path -> anchors array
    let docs = [];

    const renderNoResults = (msg) => {
        listContainer.innerHTML = '';
        const noRes = document.createElement('div');
        noRes.className = 'no-results';
        noRes.textContent = window.i18n ? window.i18n.t(msg) || msg : msg;
        listContainer.appendChild(noRes);
    };

    // Render document list
    const renderDocList = (query = '') => {
        currentDocument = null;
        backBtn.style.display = 'none';
        titleEl.textContent = '';
        listContainer.innerHTML = '';
        // Reset search field & placeholder for document mode if coming from anchor mode
        searchInput.placeholder = 'Search document…';

        const filtered = query ? docs.filter(doc =>
            includesIgnoreCase(doc.title || '', query) || includesIgnoreCase(doc.path, query))
            : docs;

        if (filtered.length === 0) {
            renderNoResults('docpicker.no_results');
            return;
        }

        filtered.forEach(doc => {
            const btn = createDocButton(doc, async () => {
                try {
                    let anchors = anchorsCache[doc.path];
                    if (!anchors) {
                        const md = await fetchDocumentSource(doc.path);
                        anchors = extractAnchors(md);
                        anchorsCache[doc.path] = anchors;
                    }
                    renderAnchorList(doc, anchors);
                } catch (err) {
                    console.error(err);
                    renderNoResults('Failed to load headings');
                }
            });
            listContainer.appendChild(btn);
        });
    };

    // Render anchor list for a selected document
    const renderAnchorList = (doc, anchors, query = '') => {
        currentDocument = doc;
        backBtn.style.display = 'inline-block';
        titleEl.textContent = doc.title || doc.path;
        listContainer.innerHTML = '';
        // Switch placeholder to heading search and clear existing text (optional)
        searchInput.placeholder = 'Search heading…';

        const filteredAnchors = query ? anchors.filter(a =>
            includesIgnoreCase(a.text, query) || includesIgnoreCase(a.id, query))
            : anchors;

        if (filteredAnchors.length === 0) {
            renderNoResults('No headings found');
            return;
        }

        filteredAnchors.forEach(a => {
            const btn = document.createElement('button');
            btn.className = `anchor-btn level-${a.level}`;
            btn.textContent = `${'  '.repeat(a.level - 1)}${a.text}`;
            btn.title = `#${a.id}`;
            btn.addEventListener('click', (ev) => {
                ev.stopPropagation();
                insertDocAnchorLink(doc, a);
                hideAnchorPicker();
            });
            listContainer.appendChild(btn);
        });
    };

    // Back button behavior
    backBtn.addEventListener('click', () => {
        searchInput.value = '';
        renderDocList();
    });

    // Search handler
    searchInput.addEventListener('input', (e) => {
        const query = e.target.value;
        if (currentDocument) {
            const anchors = anchorsCache[currentDocument.path] || [];
            renderAnchorList(currentDocument, anchors, query);
        } else {
            renderDocList(query);
        }
    });

    // Initial fetch of documents (reuse global cache if available)
    (async () => {
        try {
            docs = documentData.length ? documentData : await fetchDocuments();
            renderDocList();
        } catch (err) {
            renderNoResults(err.message || window.i18n ? window.i18n.t('docpicker.error_loading') : 'Error loading documents');
        }
    })();

    // Expose refresh method for showAnchorPicker()
    picker.renderList = () => {
        if (currentDocument) {
            const anchors = anchorsCache[currentDocument.path] || [];
            renderAnchorList(currentDocument, anchors, searchInput.value || '');
        } else {
            renderDocList(searchInput.value || '');
        }
    };

    // Expose a reset method so the caller can force picker to start at doc list
    picker.resetView = () => {
        searchInput.value = '';
        renderDocList();
    };

    return picker;
}

// Show anchor picker
function showAnchorPicker(button) {
    if (!anchorPickerElement) {
        anchorPickerElement = createDocAnchorPicker();
    }

    // Always reset to the top-level document list before showing
    if (typeof anchorPickerElement.resetView === 'function') {
        anchorPickerElement.resetView();
    }

    if (anchorPickerElement.style.display === 'block') {
        hideAnchorPicker();
        return;
    }

    // Refresh headings list each time the picker is shown to capture any new headings
    if (typeof anchorPickerElement.renderList === 'function') {
        const queryInput = anchorPickerElement.querySelector('.anchor-search-input');
        const currentQuery = queryInput ? queryInput.value : '';
        anchorPickerElement.renderList(currentQuery);
    }

    anchorPickerElement.style.display = 'block';
    requestAnimationFrame(() => {
        positionPicker(anchorPickerElement, button);
        setTimeout(() => positionPicker(anchorPickerElement, button), 50);
    });

    setTimeout(() => {
        const input = anchorPickerElement.querySelector('.anchor-search-input');
        if (input) input.focus();
    }, 50);

    const closeHandler = function(e) {
        if (!anchorPickerElement.contains(e.target) &&
            !e.target.closest('.anchor-link-button')) {
            hideAnchorPicker();
        }
    };

    // Store and register the handler so it can be cleaned up later
    anchorPickerElement._closeHandler = closeHandler;
    setTimeout(() => document.addEventListener('click', closeHandler), 0);
}

function hideAnchorPicker() {
    if (anchorPickerElement) {
        anchorPickerElement.style.display = 'none';

        // Clean up any outstanding outside-click listener
        if (anchorPickerElement._closeHandler) {
            document.removeEventListener('click', anchorPickerElement._closeHandler);
            anchorPickerElement._closeHandler = null;
        }
    }
}

// ===== UTILITY FUNCTIONS =====

// Case-insensitive substring check
function includesIgnoreCase(value = '', query = '') {
    return value.toLowerCase().includes((query || '').toLowerCase());
}

// Factory that returns a button element (used by both doc-link and anchor pickers)
function createDocButton(doc, onSelect) {
    const btn = document.createElement('button');
    btn.className = 'doc-btn';
    btn.title = doc.path;
    btn.style.boxSizing = 'border-box';
    btn.style.width = '100%';

    const docName = document.createElement('div');
    docName.className = 'doc-name';
    docName.textContent = doc.title || doc.path.split('/').pop();

    const docPath = document.createElement('div');
    docPath.className = 'doc-path';
    docPath.textContent = doc.path;

    btn.appendChild(docName);
    btn.appendChild(docPath);

    btn.addEventListener('click', (e) => {
        e.stopPropagation();
        if (typeof onSelect === 'function') onSelect();
    });

    return btn;
}

// Add helper to position picker elements based on real size instead of a hard-coded estimate
function positionPicker(picker, button) {
    // The picker must already be visible (display:block) before calling this
    const rect = button.getBoundingClientRect();
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    const pickerWidth = picker.offsetWidth;
    const pickerHeight = picker.offsetHeight;

    // Initial position – directly under the button
    let left = rect.left;
    let top = rect.bottom;

    const isSmallScreen = viewportWidth < 500;

    if (isSmallScreen) {
        // Centre horizontally on very small screens
        left = Math.max(10, (viewportWidth - pickerWidth) / 2);
        // If the button sits in the lower half, place picker above it
        if (rect.bottom > viewportHeight / 2) {
            top = rect.top - pickerHeight - 5;
        }
    } else {
        // Ensure the picker does not overflow the right edge
        if (left + pickerWidth > viewportWidth - 10) {
            // Align the right edges of button and picker
            left = rect.right - pickerWidth;
            // Keep at least 10 px padding from the viewport's left side
            left = Math.max(10, left);
        }
        // Ensure the picker does not overflow the bottom edge
        if (top + pickerHeight > viewportHeight - 10) {
            // Place the picker above the button
            top = rect.top - pickerHeight - 5;
            // Keep at least 10 px from the top edge
            if (top < 10) {
                top = 10;
            }
        }
    }

    picker.style.left = `${left}px`;
    picker.style.top = `${top}px`;
}

// Generic helper to toggle pickers (emoji, table, doc, etc.)
function togglePicker({ get, set, create, button, closeSelector, onShow, onHide, beforeShow }) {
    let picker = get();

    // Lazily create picker if not existing
    if (!picker) {
        picker = create();
        if (typeof set === 'function') set(picker);
    }

    // If picker is currently visible, hide it and exit
    if (picker.style.display === 'block') {
        picker.style.display = 'none';
        picker.style.transform = '';
        if (picker._closeHandler) {
            document.removeEventListener('click', picker._closeHandler);
            picker._closeHandler = null;
        }
        if (typeof onHide === 'function') onHide(picker);
        return;
    }

    if (typeof beforeShow === 'function') beforeShow(picker);

    // Show the picker
    picker.style.display = 'block';

    // Position after it's visible
    requestAnimationFrame(() => {
        positionPicker(picker, button);
        setTimeout(() => positionPicker(picker, button), 50);
    });

    if (typeof onShow === 'function') onShow(picker);

    // Outside click handler
    const closeHandler = function(e) {
        if (!picker.contains(e.target) && !e.target.closest(closeSelector)) {
            picker.style.display = 'none';
            picker.style.transform = '';
            document.removeEventListener('click', closeHandler);
            picker._closeHandler = null;
            if (typeof onHide === 'function') onHide(picker);
        }
    };

    // Delay registration to avoid immediate close
    setTimeout(() => document.addEventListener('click', closeHandler), 0);
    picker._closeHandler = closeHandler;
}

// Cleanup function to remove all pickers
function cleanup() {
    // Remove table picker if it exists
    if (tablePickerElement) {
        tablePickerElement.remove();
        tablePickerElement = null;
    }

    // Remove emoji picker if it exists
    if (emojiPickerElement) {
        emojiPickerElement.remove();
        emojiPickerElement = null;
    }

    // Remove doc picker if it exists
    if (docPickerElement) {
        docPickerElement.remove();
        docPickerElement = null;
    }

    // Remove anchor picker if it exists
    if (anchorPickerElement) {
        anchorPickerElement.remove();
        anchorPickerElement = null;
    }
}

// Preload emoji data when the page loads
document.addEventListener('DOMContentLoaded', function() {
    // Preload emoji data in the background - browsers will also preload it via Link header
    setTimeout(() => {
        EmojiCache.getData().then(() => {
            console.log('Emoji data preloaded successfully');
        });
    }, 1000); // Delay by 1 second to let page finish loading first
});

// Export the module
window.EditorPickers = {
    // Emoji picker
    showEmojiPicker,
    hideEmojiPicker,
    
    // Document picker
    showDocPicker,
    hideDocPicker,
    
    // Table picker
    showTablePicker,
    
    // Anchor picker
    showAnchorPicker,
    hideAnchorPicker,
    
    // Utility
    cleanup,
    EmojiCache
};