/**
 * Clipboard handling module for Wiki-Go
 * Enables pasting images from clipboard directly into the editor and URL link creation
 */

// Main initialization function for clipboard handling
function initClipboardHandling() {
    console.log('Initializing clipboard handler');

    // Handle clipboard paste events at document level (for images)
    document.addEventListener('paste', documentPasteHandler);

    // Check if link embedding from clipboard is enabled
    if (window.WikiConfig && window.WikiConfig.enableLinkEmbedding) {
        // Add direct binding to CodeMirror textarea (for URL links)
        setupDirectTextareaBinding();
    } else {
        console.log('Link embedding from clipboard is disabled');
    }
}

/**
 * Set up direct binding to CodeMirror textarea for URL pasting
 */
function setupDirectTextareaBinding() {
    // Bind to any existing CodeMirror instances
    bindToExistingCodeMirrors();

    // Set up observer to detect new CodeMirror instances
    const observer = new MutationObserver(mutations => {
        for (const mutation of mutations) {
            if (mutation.addedNodes.length) {
                bindToExistingCodeMirrors();
                break;
            }
        }
    });

    // Start observing the DOM for changes
    observer.observe(document.body, { childList: true, subtree: true });

    // Check again after delays (for dynamically loaded editors)
    setTimeout(bindToExistingCodeMirrors, 1000);
    setTimeout(bindToExistingCodeMirrors, 2000);
}

/**
 * Find and bind to CodeMirror textarea elements
 */
function bindToExistingCodeMirrors() {
    // Find all CodeMirror textareas
    const textareas = document.querySelectorAll('.CodeMirror textarea');
    if (textareas.length === 0) {
        return;
    }

    textareas.forEach(textarea => {
        // Only bind once to each textarea
        if (!textarea.hasAttribute('data-url-paste-bound')) {
            // Add paste event listener with capture to get it before CodeMirror
            textarea.addEventListener('paste', handleUrlPaste, true);

            // Mark as bound
            textarea.setAttribute('data-url-paste-bound', 'true');
            console.log('URL paste handler bound to CodeMirror textarea');
        }
    });
}

/**
 * Handle URL paste directly on the CodeMirror textarea
 * @param {ClipboardEvent} event - The paste event
 */
function handleUrlPaste(event) {
    // Find the CodeMirror instance
    const cmElement = event.target.closest('.CodeMirror');
    if (!cmElement || !cmElement.CodeMirror) {
        return;
    }

    const editor = cmElement.CodeMirror;

    // Check if text is selected
    if (!editor.somethingSelected()) {
        return;
    }

    // Get clipboard text
    const clipboardText = event.clipboardData.getData('text/plain').trim();
    if (!clipboardText) {
        return; // Empty clipboard text
    }

    // Check if it looks like a URL (permissive check)
    const looksLikeUrl = clipboardText.includes('//') ||
                         clipboardText.includes('www.') ||
                         clipboardText.includes('.') ||
                         clipboardText.includes(':');

    if (looksLikeUrl) {
        // Prevent default paste behavior
        event.preventDefault();
        event.stopPropagation();

        // Get selected text
        const selectedText = editor.getSelection();

        // Check if the selection is multiline
        const isMultiLine = selectedText.includes('\n');

        if (isMultiLine) {
            // For multiline selections, just replace with the URL
            editor.replaceSelection(clipboardText);
            console.log('Replaced multiline selection with URL');
        } else {
            // For single line selections, create a markdown link

            // Format URL (add http:// if needed)
            let url = clipboardText;
            if (!url.startsWith('http://') && !url.startsWith('https://')) {
                url = 'http://' + url;
            }

            // Create markdown link
            const markdownLink = `[${selectedText}](${url})`;

            // Replace selection with markdown link
            editor.replaceSelection(markdownLink);
            console.log('Created link:', markdownLink);
        }
    }
}

/**
 * Document-level paste event handler
 * @param {ClipboardEvent} event - The paste event
 */
async function documentPasteHandler(event) {
    // Check if editor is active
    if (!window.WikiEditor || !window.WikiEditor.isEditorActive()) {
        return;
    }

    // Check if the paste event is inside the editor
    const target = event.target;
    const isInEditor = target.classList.contains('CodeMirror') ||
                      target.closest('.CodeMirror') !== null;

    if (!isInEditor) {
        return; // Not pasting in the editor, allow default behavior
    }

    // Handle paste event for images
    await handleImagePaste(event);
}

/**
 * Handle image paste events in the editor
 * @param {ClipboardEvent} event - The paste event
 */
async function handleImagePaste(event) {
    // Check if there are any clipboard items
    if (!event.clipboardData || !event.clipboardData.items) {
        return;
    }

    // Get the editor instance
    const cmElement = document.querySelector('.CodeMirror');
    if (!cmElement || !cmElement.CodeMirror) {
        return;
    }
    const editor = cmElement.CodeMirror;

    // Look for images in clipboard data
    const items = event.clipboardData.items;
    let imageFile = null;

    for (let i = 0; i < items.length; i++) {
        if (items[i].type.startsWith('image/')) {
            // Found an image, get the file
            imageFile = items[i].getAsFile();
            break;
        }
    }

    // If no image was found, allow default paste behavior
    if (!imageFile) {
        return;
    }

    // Prevent default paste behavior for images
    event.preventDefault();

    try {
        // Get the current document path
        const docPath = getCurrentDocPath();

        // Generate a timestamp-based filename with the correct extension
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const extension = imageFile.type.split('/')[1] || 'png';
        const filename = `image-${timestamp}.${extension}`;

        // Create FormData to upload the image
        const formData = new FormData();
        formData.append('file', imageFile, filename);
        formData.append('docPath', docPath);

        // Show a temporary loading message at cursor
        const loadingPlaceholder = `![Uploading ${filename}...]()`;
        const cursor = editor.getCursor();
        editor.replaceRange(loadingPlaceholder, cursor);

        // Upload the image
        const response = await fetch(wikiURL('/api/files/upload'), {
            method: 'POST',
            body: formData
        });

        // Process the response
        const data = await response.json();

        // Find the position of the loading placeholder
        const text = editor.getValue();
        const placeholderPos = text.indexOf(loadingPlaceholder);

        if (data.success) {
            // Success - replace the placeholder with the actual image markdown
            const imageMarkdown = `![](${filename})`;

            if (placeholderPos >= 0) {
                const placeholderStart = editor.posFromIndex(placeholderPos);
                const placeholderEnd = editor.posFromIndex(placeholderPos + loadingPlaceholder.length);
                editor.replaceRange(imageMarkdown, placeholderStart, placeholderEnd);
            } else {
                // If placeholder can't be found for some reason, insert at current cursor
                editor.replaceSelection(imageMarkdown);
            }

            console.log('Image uploaded successfully:', filename);
        } else {
            // Error - replace the placeholder with an error message
            const errorMessage = `<!-- Failed to upload image: ${data.message || 'Unknown error'} -->`;

            if (placeholderPos >= 0) {
                const placeholderStart = editor.posFromIndex(placeholderPos);
                const placeholderEnd = editor.posFromIndex(placeholderPos + loadingPlaceholder.length);
                editor.replaceRange(errorMessage, placeholderStart, placeholderEnd);
            } else {
                // If placeholder can't be found, insert at current cursor
                editor.replaceSelection(errorMessage);
            }

            console.error('Failed to upload image:', data.message);
        }
    } catch (error) {
        console.error('Error handling clipboard paste:', error);
        // Try to get the editor instance again for error handling
        try {
            const cmElement = document.querySelector('.CodeMirror');
            if (cmElement && cmElement.CodeMirror) {
                cmElement.CodeMirror.replaceSelection(`<!-- Error uploading pasted image: ${error.message} -->`);
            }
        } catch (e) {
            console.error('Could not insert error message into editor:', e);
        }
    }
}

// Initialize clipboard handling when the page loads
document.addEventListener('DOMContentLoaded', initClipboardHandling);

// Expose functions globally if needed
window.ClipboardHandler = {
    init: initClipboardHandling,
    handleUrlPaste: handleUrlPaste,
    handleImagePaste: handleImagePaste
};