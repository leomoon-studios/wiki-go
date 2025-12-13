/**
 * Backup Manager Module
 * Handles backup creation and management
 */

document.addEventListener('DOMContentLoaded', function() {
    'use strict';

    // Elements
    const createBackupBtn = document.getElementById('createBackupBtn');
    const backupProgressBar = document.getElementById('backupProgressBar');
    const backupProgressText = document.getElementById('backupProgressText');
    const backupProgressDetails = document.getElementById('backupProgressDetails');
    const backupProgressContainer = document.querySelector('.backup-progress-container');
    const backupList = document.getElementById('backupList');
    const backupTabBtn = document.querySelector('button[data-tab="backup-tab"]');

    // Initialize
    if (backupTabBtn) {
        backupTabBtn.addEventListener('click', loadBackups);
    }

    if (createBackupBtn) {
        createBackupBtn.addEventListener('click', startBackup);
    }

    // Functions
    async function loadBackups() {
        if (!backupList) return;

        backupList.innerHTML = `<div class="empty-message">${window.i18n ? window.i18n.t('backup.loading') : 'Loading backups...'}</div>`;

        try {
            const response = await fetch('/api/backup/list');
            if (response.ok) {
                const data = await response.json();
                renderBackups(data.backups || []);
            } else {
                backupList.innerHTML = `<div class="error-message">${window.i18n ? window.i18n.t('backup.error_loading') : 'Failed to load backups'}</div>`;
            }
        } catch (error) {
            console.error('Error loading backups:', error);
            backupList.innerHTML = `<div class="error-message">${window.i18n ? window.i18n.t('backup.error_loading') : 'Failed to load backups'}</div>`;
        }
    }

    function renderBackups(backups) {
        if (!backupList) return;
        
        backupList.innerHTML = '';
        
        if (backups.length === 0) {
            backupList.innerHTML = `<div class="empty-message">${window.i18n ? window.i18n.t('backup.no_backups') : 'No backups found'}</div>`;
            return;
        }

        backups.forEach(backup => {
            const item = document.createElement('div');
            item.className = 'file-item';
            
            const sizeFormatted = formatBytes(backup.size);
            
            item.innerHTML = `
                <div class="file-info">
                    <div class="file-icon"><i class="fa fa-file-zip-o"></i></div>
                    <div class="file-details" style="display: flex; flex-direction: column; overflow: hidden;">
                        <span class="file-name" title="${backup.name}">${backup.name}</span>
                        <span class="file-meta" style="font-size: 0.85em; color: var(--text-muted);">${backup.date} • ${sizeFormatted}</span>
                    </div>
                </div>
                <div class="file-actions">
                    <a href="${backup.url}" class="download-file-btn" title="${window.i18n ? window.i18n.t('common.download') : 'Download'}" download>
                        <i class="fa fa-download"></i>
                    </a>
                    <button class="delete-file-btn" data-filename="${backup.name}" title="${window.i18n ? window.i18n.t('common.delete') : 'Delete'}">
                        <i class="fa fa-trash"></i>
                    </button>
                </div>
            `;
            
            const deleteBtn = item.querySelector('.delete-file-btn');
            if (deleteBtn) {
                deleteBtn.onclick = () => deleteBackup(backup.name);
            }
            
            backupList.appendChild(item);
        });
    }

    async function startBackup() {
        if (createBackupBtn.disabled) return;
        
        createBackupBtn.disabled = true;
        showProgress();

        try {
            const response = await fetch('/api/backup/start', { method: 'POST' });
            
            if (!response.ok) {
                throw new Error('Failed to start backup');
            }

            const data = await response.json();
            if (data.statusUrl) {
                pollStatus(data.statusUrl);
            } else {
                throw new Error('No status URL returned');
            }
        } catch (error) {
            console.error('Backup error:', error);
            window.DialogSystem.showMessageDialog(
                window.i18n ? window.i18n.t('common.error') : 'Error',
                window.i18n ? window.i18n.t('backup.error_start') : 'Failed to start backup'
            );
            resetProgress();
            createBackupBtn.disabled = false;
        }
    }

    async function pollStatus(url) {
        try {
            const response = await fetch(url);
            if (response.ok) {
                const status = await response.json();
                
                updateProgress(status.progress, status.currentFile);

                if (status.status === 'completed') {
                    finishBackup();
                } else if (status.status === 'failed') {
                    throw new Error(status.error || 'Backup failed');
                } else {
                    setTimeout(() => pollStatus(url), 1000);
                }
            } else {
                throw new Error('Failed to get status');
            }
        } catch (error) {
            console.error('Poll error:', error);
            window.DialogSystem.showMessageDialog(
                window.i18n ? window.i18n.t('common.error') : 'Error',
                (window.i18n ? window.i18n.t('backup.error_failed') : 'Backup failed') + ': ' + error.message
            );
            resetProgress();
            createBackupBtn.disabled = false;
        }
    }

    function finishBackup() {
        updateProgress(100, window.i18n ? window.i18n.t('backup.completed') : 'Backup completed');
        setTimeout(() => {
            resetProgress();
            createBackupBtn.disabled = false;
            loadBackups();
        }, 1000);
    }

    function deleteBackup(filename) {
        const title = window.i18n ? window.i18n.t('backup.delete_title') : 'Delete Backup';
        const message = window.i18n ? window.i18n.t('backup.confirm_delete') : 'Are you sure you want to delete this backup?';

        window.DialogSystem.showConfirmDialog(title, message, async (confirmed) => {
            if (!confirmed) return;

            try {
                const response = await fetch(`/api/backup/delete/${filename}`, { method: 'DELETE' });
                if (response.ok) {
                    loadBackups();
                } else {
                    window.DialogSystem.showMessageDialog(
                        window.i18n ? window.i18n.t('common.error') : 'Error',
                        window.i18n ? window.i18n.t('backup.error_delete') : 'Failed to delete backup'
                    );
                }
            } catch (error) {
                console.error('Delete error:', error);
                window.DialogSystem.showMessageDialog(
                    window.i18n ? window.i18n.t('common.error') : 'Error',
                    window.i18n ? window.i18n.t('backup.error_delete') : 'Failed to delete backup'
                );
            }
        });
    }

    function showProgress() {
        if (backupProgressContainer) backupProgressContainer.style.display = 'block';
        updateProgress(0, window.i18n ? window.i18n.t('backup.starting') : 'Starting...');
    }

    function resetProgress() {
        if (backupProgressContainer) backupProgressContainer.style.display = 'none';
        updateProgress(0, '');
    }

    function updateProgress(percent, text) {
        if (backupProgressBar) backupProgressBar.style.width = `${percent}%`;
        if (backupProgressText) backupProgressText.textContent = `${percent}%`;
        if (backupProgressDetails) backupProgressDetails.textContent = text || '';
    }

    function formatBytes(bytes, decimals = 2) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const dm = decimals < 0 ? 0 : decimals;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
    }
});
