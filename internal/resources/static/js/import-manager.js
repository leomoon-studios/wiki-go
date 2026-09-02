/**
 * Import Manager Module
 * Handles document import functionality for admin users
 */

document.addEventListener('DOMContentLoaded', function() {
    'use strict';

    // Import form elements
    const importForm = document.getElementById('importForm');
    const importZipFile = document.getElementById('importZipFile');
    const importButton = document.getElementById('importButton');
    const cancelImportButton = document.getElementById('cancelImportButton');
    const importProgressContainer = document.querySelector('.import-progress-container');
    const importProgressBar = document.getElementById('importProgressBar');
    const importProgressText = document.getElementById('importProgressText');
    const importProgressDetails = document.getElementById('importProgressDetails');
    const importResults = document.querySelector('.import-results');
    const importResultsContent = document.getElementById('importResults');

    // Initialize import functionality if form exists
    if (importForm) {
        importForm.addEventListener('submit', handleImportSubmit);
    }

    if (cancelImportButton) {
        cancelImportButton.addEventListener('click', resetImportForm);
    }

    if (importZipFile) {
        importZipFile.addEventListener('change', function() {
            // Enable/disable import button based on file selection
            if (importButton) {
                importButton.disabled = !importZipFile.files.length;
            }
        });
    }

    /**
     * Handle import form submission
     * @param {Event} e - Form submit event
     */
    async function handleImportSubmit(e) {
        e.preventDefault();

        // Validate file selection
        if (!importZipFile.files.length) {
            showImportError('Please select a ZIP file to import');
            return;
        }

        const file = importZipFile.files[0];

        // Validate file type
        if (!file.name.toLowerCase().endsWith('.zip')) {
            showImportError('Please select a valid ZIP file');
            return;
        }

        // Create form data
        const formData = new FormData();
        formData.append('zipFile', file);

        try {
            // Show progress UI
            showImportProgress();

            // Start the import process
            const response = await fetch('/api/import', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || 'Import failed');
            }

            const data = await response.json();

            // Get the job status URL
            if (data.statusUrl) {
                // Poll for status updates
                pollImportStatus(data.statusUrl);
            } else {
                throw new Error('No status URL provided');
            }
        } catch (error) {
            console.error('Import error:', error);
            resetImportProgress();
            showImportError(error.message || 'Failed to start import process');
        }
    }

    /**
     * Poll the import status endpoint for updates
     * @param {string} statusUrl - URL to check import status
     */
    async function pollImportStatus(statusUrl) {
        try {
            const response = await fetch(statusUrl);

            if (!response.ok) {
                throw new Error('Failed to get import status');
            }

            const data = await response.json();

            // Update progress
            updateImportProgress(data);

            // If import is still in progress, poll again after a delay
            if (data.status === 'processing') {
                setTimeout(() => pollImportStatus(statusUrl), 1000);
            } else if (data.status === 'completed') {
                // Show completion message and results
                showImportResults(data);
            } else if (data.status === 'failed') {
                // Show error message
                showImportError(data.message || 'Import failed');
                resetImportProgress();
            }
        } catch (error) {
            console.error('Error polling import status:', error);
            showImportError('Failed to get import status');
            resetImportProgress();
        }
    }

    /**
     * Update the import progress UI
     * @param {Object} data - Import status data
     */
    function updateImportProgress(data) {
        if (!data) return;

        const progress = data.progress || 0;

        // Update progress bar
        if (importProgressBar) {
            importProgressBar.style.width = `${progress}%`;
        }

        // Update progress text
        if (importProgressText) {
            importProgressText.textContent = `${progress}%`;
        }

        // Update progress details
        if (importProgressDetails && data.currentFile) {
            importProgressDetails.textContent = `Processing: ${data.currentFile}`;
        }
    }

    /**
     * Show the import progress UI
     */
    function showImportProgress() {
        if (importButton) {
            importButton.disabled = true;
            importButton.textContent = window.i18n ? window.i18n.t('import.importing') : 'Importing...';
        }

        if (importZipFile) {
            importZipFile.disabled = true;
        }

        if (importProgressContainer) {
            importProgressContainer.style.display = 'block';
        }

        // Hide results if they were previously shown
        if (importResults) {
            importResults.style.display = 'none';
        }
    }

    /**
     * Reset the import progress UI
     */
    function resetImportProgress() {
        if (importButton) {
            importButton.disabled = false;
            importButton.textContent = window.i18n ? window.i18n.t('import.start_button') : 'Import';
        }

        if (importZipFile) {
            importZipFile.disabled = false;
        }

        if (importProgressContainer) {
            importProgressContainer.style.display = 'none';
        }

        // Reset progress bar
        if (importProgressBar) {
            importProgressBar.style.width = '0%';
        }

        // Reset progress text
        if (importProgressText) {
            importProgressText.textContent = '0%';
        }

        // Reset progress details
        if (importProgressDetails) {
            importProgressDetails.textContent = '';
        }
    }

    /**
     * Show import results
     * @param {Object} data - Import results data
     */
    function showImportResults(data) {
        resetImportProgress();

        if (!importResults || !importResultsContent) return;

        // Show results container
        importResults.style.display = 'block';

        const fragment = document.createDocumentFragment();

        if (data.importedFiles && data.importedFiles.length) {
            const importedList = WikiDOM.element('ul', 'imported-files-list');

            data.importedFiles.forEach(file => {
                const item = document.createElement('li');
                item.appendChild(document.createTextNode(`${file.originalPath ?? ''} → `));
                const link = WikiDOM.element('a', '', file.newPath);
                link.href = WikiDOM.localURL(file.newPath);
                link.target = '_blank';
                link.rel = 'noopener';
                item.appendChild(link);
                importedList.appendChild(item);
            });

            fragment.appendChild(importedList);

            // Refresh the sidebar to show new content
            if (window.SidebarNavigation && window.SidebarNavigation.refreshSidebar) {
                window.SidebarNavigation.refreshSidebar();
            }
        }

        if (data.errors && data.errors.length) {
            fragment.appendChild(WikiDOM.element('h5', '', 'Errors:'));
            const errorsList = WikiDOM.element('ul', 'import-errors-list');

            data.errors.forEach(error => {
                errorsList.appendChild(WikiDOM.element('li', '', error));
            });

            fragment.appendChild(errorsList);
        }

        // Add summary
        fragment.appendChild(WikiDOM.element('p', 'import-summary', `Successfully imported ${data.successCount || 0} files with ${data.errorCount || 0} errors.`));

        // Update results content
        importResultsContent.replaceChildren(fragment);
    }

    /**
     * Show import error message
     * @param {string} message - Error message to display
     */
    function showImportError(message) {
        const errorMessage = document.querySelector('.settings-dialog .error-message');

        if (errorMessage) {
            errorMessage.textContent = message;
            errorMessage.style.display = 'block';
        } else {
            // Fallback to alert if error message element not found
            alert(message);
        }
    }

    /**
     * Reset the import form
     */
    function resetImportForm() {
        if (importForm) {
            importForm.reset();
        }

        resetImportProgress();

        // Hide results
        if (importResults) {
            importResults.style.display = 'none';
        }

        // Clear error message
        const errorMessage = document.querySelector('.settings-dialog .error-message');
        if (errorMessage) {
            errorMessage.style.display = 'none';
        }

        // Disable import button
        if (importButton) {
            importButton.disabled = true;
        }
    }

    // Export functions for global access
    window.ImportManager = {
        resetImportForm
    };
});
