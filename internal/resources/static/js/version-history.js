// Version History functionality for Wiki-Go
(function() {
    'use strict';

    // Global references to DOM elements
    let versionHistoryDialog;
    let closeVersionHistoryDialog;
    let versionList;
    let versionPreview;

    // Initialize elements when the DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        versionHistoryDialog = document.querySelector('.version-history-dialog');
        closeVersionHistoryDialog = versionHistoryDialog?.querySelector('.close-dialog');
        versionList = document.querySelector('.version-list');
        versionPreview = document.querySelector('.version-preview-container');

        // Event listeners for version history dialog
        const viewHistoryButton = document.querySelector('.view-history');
        if (viewHistoryButton) {
            viewHistoryButton.addEventListener('click', function() {
                showVersionHistoryDialog();
            });
        }

        if (closeVersionHistoryDialog) {
            closeVersionHistoryDialog.addEventListener('click', hideVersionHistoryDialog);
        }

        // Escape key is now handled by keyboard-shortcuts.js
    });

    // Use the shared getCurrentDocPath function from utilities.js

    // Show version history dialog
    function showVersionHistoryDialog() {
        console.log("Opening version history dialog");
        try {
            // Ensure close button is visible
            if (closeVersionHistoryDialog) {
                closeVersionHistoryDialog.style.display = 'flex';
            }

            versionHistoryDialog.classList.add('active');
            // Load the document versions
            loadDocumentVersions();
        } catch (error) {
            console.error("Error opening version history dialog:", error);
            alert("An error occurred opening the version history dialog. See console for details.");
        }
    }

    // Hide version history dialog
    function hideVersionHistoryDialog() {
        console.log("Closing version history dialog");
        try {
            versionHistoryDialog.classList.remove('active');
            // Reset preview content
            const previewElement = document.querySelector('.version-preview');
            if (previewElement) {
                const message = window.i18n ? window.i18n.t('history.select_version') : 'Select a version to preview';
                WikiDOM.message(previewElement, 'empty-message', message);
            }
        } catch (error) {
            console.error("Error closing version history dialog:", error);
        }
    }

    // Load document versions from the server
    async function loadDocumentVersions() {
        const path = getCurrentDocPath();
        console.log("Loading versions for document path:", path);
        WikiDOM.message(versionList, 'loading-spinner', 'Loading versions...');

        try {
            const apiUrl = `/api/versions/${path}`;
            console.log("Requesting versions from:", apiUrl);

            const response = await fetch(apiUrl);
            console.log("API response status:", response.status);

            if (!response.ok) {
                throw new Error(`Failed to load versions: ${response.status}`);
            }

            const data = await response.json();
            console.log("API response data:", data);

            if (!data.success) {
                throw new Error(data.message || 'Failed to load document versions');
            }

            // Render the versions list
            console.log("Number of versions found:", data.versions ? data.versions.length : 0);
            renderVersionsList(data.versions);
        } catch (error) {
            console.error('Error loading document versions:', error);
            WikiDOM.message(versionList, 'error-message', `Failed to load versions: ${error.message}`);
        }
    }

    // Render the list of document versions
    function renderVersionsList(versions) {
        if (!versions || versions.length === 0) {
            WikiDOM.message(versionList, 'empty-message', window.i18n ? window.i18n.t('history.no_versions') : 'No previous versions found');
            return;
        }

        const fragment = document.createDocumentFragment();
        versions.forEach(version => {
            // Create a Date object from the version's timestamp (format: yyyymmddhhmmss)
            const timestamp = String(version?.timestamp ?? '');
            if (!/^\d{14}$/.test(timestamp)) return;
            const year = timestamp.substring(0, 4);
            const month = timestamp.substring(4, 6);
            const day = timestamp.substring(6, 8);
            const hour = timestamp.substring(8, 10);
            const minute = timestamp.substring(10, 12);
            const second = timestamp.substring(12, 14);

            const date = new Date(`${year}-${month}-${day}T${hour}:${minute}:${second}`);
            const formattedDate = date.toLocaleString();

            const item = WikiDOM.element('div', 'version-item');
            item.dataset.version = String(version.timestamp ?? '');
            const info = WikiDOM.element('div', 'version-info');
            info.appendChild(WikiDOM.element('div', 'version-date', formattedDate));

            const actions = WikiDOM.element('div', 'version-actions');
            const preview = WikiDOM.element('button', 'preview-version-btn');
            preview.type = 'button';
            preview.title = window.i18n ? window.i18n.t('history.preview_button') : 'Preview this version';
            preview.dataset.i18nTitle = 'history.preview_button';
            const previewLabel = WikiDOM.element('span', '', window.i18n ? window.i18n.t('history.preview_button') : 'Preview');
            previewLabel.dataset.i18n = 'history.preview_button';
            preview.append(WikiDOM.icon('fa fa-eye'), previewLabel);
            const restore = WikiDOM.element('button', 'restore-version-btn');
            restore.type = 'button';
            restore.title = window.i18n ? window.i18n.t('history.restore_button') : 'Restore this version';
            restore.dataset.i18nTitle = 'history.restore_button';
            const restoreLabel = WikiDOM.element('span', '', window.i18n ? window.i18n.t('history.restore_button') : 'Restore');
            restoreLabel.dataset.i18n = 'history.restore_button';
            restore.append(WikiDOM.icon('fa fa-history'), restoreLabel);
            actions.append(preview, restore);
            item.append(info, actions);
            fragment.appendChild(item);
        });

        versionList.replaceChildren(fragment);

        // Add event listeners for version actions
        versionList.querySelectorAll('.preview-version-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const versionItem = e.target.closest('.version-item');
                const version = versionItem.getAttribute('data-version');
                previewVersion(version);

                // Highlight the selected version
                versionList.querySelectorAll('.version-item').forEach(item => {
                    item.classList.remove('selected');
                });
                versionItem.classList.add('selected');
            });
        });

        versionList.querySelectorAll('.restore-version-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const version = e.target.closest('.version-item').getAttribute('data-version');
                confirmRestoreVersion(version);
            });
        });
    }

    // Preview a specific version
    async function previewVersion(version) {
        const path = getCurrentDocPath();
        const previewContainer = document.querySelector('.version-preview-container');
        const previewElement = document.querySelector('.version-preview');

        console.log("Preview container:", previewContainer);
        console.log("Preview element:", previewElement);

        // Determine which element to use for the preview content
        const targetElement = previewElement || previewContainer;
        WikiDOM.message(targetElement, 'loading-spinner', 'Loading preview...');

        try {
            // First, fetch the raw content of the version
            const response = await fetch(`/api/versions/${path}/${version}`);
            console.log("Preview response status:", response.status);

            if (!response.ok) {
                throw new Error(`Failed to load version: ${response.status}`);
            }

            const data = await response.json();
            console.log("Preview data received:", data);

            if (!data.success) {
                throw new Error(data.message || 'Failed to load document version');
            }

            // Now fetch the rendered HTML by requesting the content to be rendered by the server
            // This ensures we use the same rendering engine as the main content
            // Pass the document path as a query parameter so file attachments can be properly transformed
            const renderResponse = await fetch(`/api/render-markdown?path=${encodeURIComponent(path)}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
                },
                body: data.content
            });

            if (!renderResponse.ok) {
                throw new Error(`Failed to render markdown: ${renderResponse.status}`);
            }

            // Get the rendered HTML directly from the server
            const renderedHTML = await renderResponse.text();

            // Display the rendered content
            const versionContent = WikiDOM.element('div', 'version-content markdown-body');
            // SECURITY: this endpoint uses the same safe server-side Markdown renderer as document pages.
            versionContent.innerHTML = renderedHTML;
            targetElement.replaceChildren(versionContent);

            // Use lazy loader to load libraries if needed
            const promises = [];
            
            // Check and load Prism if there are code blocks
            if (targetElement.querySelector('pre code')) {
                if (window.LazyLoader) {
                    promises.push(window.LazyLoader.forceLoad('prism').then(() => {
                        if (typeof Prism !== 'undefined') {
                            targetElement.querySelectorAll('pre code').forEach((block) => {
                                Prism.highlightElement(block);
                            });
                        }
                    }));
                } else if (typeof Prism !== 'undefined') {
                    targetElement.querySelectorAll('pre code').forEach((block) => {
                        Prism.highlightElement(block);
                    });
                }
            }

            // Check and load MathJax if there are math formulas
            if (targetElement.querySelector('.math, .katex, [class*="math"]') || 
                targetElement.textContent.includes('$')) {
                if (window.LazyLoader) {
                    promises.push(window.LazyLoader.forceLoad('mathjax').then(() => {
                        if (typeof MathJax !== 'undefined') {
                            try {
                                MathJax.typesetPromise([targetElement]);
                            } catch (mathError) {
                                console.error('MathJax error:', mathError);
                            }
                        }
                    }));
                } else if (typeof MathJax !== 'undefined') {
                    try {
                        MathJax.typesetPromise([targetElement]);
                    } catch (mathError) {
                        console.error('MathJax error:', mathError);
                    }
                }
            }

            // Check and load Mermaid if there are diagrams
            if (targetElement.querySelector('.mermaid')) {
                if (window.LazyLoader) {
                    promises.push(window.LazyLoader.forceLoad('mermaid').then(() => {
                        if (typeof mermaid !== 'undefined' && window.MermaidHandler) {
                            try {
                                console.log('Using MermaidHandler for version preview');
                                window.MermaidHandler.initVersionPreview(targetElement);
                            } catch (mermaidError) {
                                console.error('Mermaid handler error:', mermaidError);
                            }
                        }
                    }));
                } else if (typeof mermaid !== 'undefined' && window.MermaidHandler) {
                    try {
                        console.log('Using MermaidHandler for version preview');
                        window.MermaidHandler.initVersionPreview(targetElement);
                    } catch (mermaidError) {
                        console.error('Mermaid handler error:', mermaidError);
                    }
                }
            }

            // Wait for all libraries to load and process
            await Promise.all(promises);
        } catch (error) {
            console.error('Error loading version preview:', error);
            WikiDOM.message(targetElement, 'error-message', `Failed to load preview: ${error.message}`);
        }
    }

    // Confirm and restore a specific version
    function confirmRestoreVersion(version) {
        window.showConfirmDialog(
            window.i18n ? window.i18n.t('restore.title') : "Restore Version",
            window.i18n ? window.i18n.t('restore.confirm_message') : "Are you sure you want to restore this version? This will replace the current document content.",
            async (confirmed) => {
                if (!confirmed) {
                    return;
                }

                try {
                    const path = getCurrentDocPath();
                    console.log(`Restoring version ${version} for document path: ${path}`);
                    const restoreUrl = `/api/versions/${path}/${version}/restore`;
                    console.log(`Sending POST request to: ${restoreUrl}`);

                    // No processing message - just send the request
                    const response = await fetch(restoreUrl, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Cache-Control': 'no-cache, no-store, must-revalidate',
                            'Pragma': 'no-cache',
                        }
                    });

                    console.log(`Restore response status: ${response.status}`);

                    let data;
                    try {
                        data = await response.json();
                        console.log("Restore response data:", data);
                    } catch (jsonError) {
                        console.error("Error parsing restore response JSON:", jsonError);
                        throw new Error("Invalid response from server");
                    }

                    if (!response.ok || !data.success) {
                        throw new Error(data.message || `Server returned ${response.status}`);
                    }

                    // Close dialogs
                    hideVersionHistoryDialog(); // Just close the version history dialog

                    // Skip showing success message and directly reload the page
                    // Create unique timestamp for cache busting
                    const timestamp = new Date().getTime();
                    // Build a new URL with cache busting parameter
                    let newUrl = window.location.pathname;
                    if (newUrl.includes('?')) {
                        newUrl = newUrl.split('?')[0];
                    }
                    newUrl += `?nocache=${timestamp}`;

                    console.log(`Reloading page with cache-busting URL: ${newUrl}`);

                    // Clear browser cache for this page if possible
                    if ('caches' in window) {
                        try {
                            caches.delete(window.location.href).then(() => {
                                console.log("Cache cleared for this page");
                            });
                        } catch (e) {
                            console.log("Could not clear cache:", e);
                        }
                    }

                    // First attempt: navigate to the URL with cache busting parameter
                    window.location.href = newUrl;

                    // Second fallback: force reload without cache
                    setTimeout(() => {
                        console.log("Fallback: using location.reload(true)");
                        window.location.reload(true);
                    }, 200);

                } catch (error) {
                    console.error('Error restoring version:', error);
                    window.showMessageDialog("Error", `Failed to restore version: ${error.message}`);
                }
            }
        );
    }

    // Expose functions to global scope
    window.VersionHistory = {
        showVersionHistoryDialog: showVersionHistoryDialog,
        hideVersionHistoryDialog: hideVersionHistoryDialog,
        loadDocumentVersions: loadDocumentVersions,
        renderVersionsList: renderVersionsList,
        previewVersion: previewVersion,
        confirmRestoreVersion: confirmRestoreVersion
    };
})();
