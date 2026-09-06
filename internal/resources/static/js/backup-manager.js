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

        WikiDOM.message(backupList, 'empty-message', window.i18n ? window.i18n.t('backup.loading') : 'Loading backups...');

        try {
            const response = await fetch('/api/backup/list');
            if (response.ok) {
                const data = await response.json();
                renderBackups(data.backups || []);
            } else {
                WikiDOM.message(backupList, 'error-message', window.i18n ? window.i18n.t('backup.error_loading') : 'Failed to load backups');
            }
        } catch (error) {
            console.error('Error loading backups:', error);
            WikiDOM.message(backupList, 'error-message', window.i18n ? window.i18n.t('backup.error_loading') : 'Failed to load backups');
        }
    }

    function renderBackups(backups) {
        if (!backupList) return;
        
        WikiDOM.clear(backupList);
        
        if (backups.length === 0) {
            WikiDOM.message(backupList, 'empty-message', window.i18n ? window.i18n.t('backup.no_backups') : 'No backups found');
            return;
        }

        backups.forEach(backup => {
            const item = document.createElement('div');
            item.className = 'file-item';
            
            const sizeFormatted = formatBytes(backup.size);
            
            const info = WikiDOM.element('div', 'file-info');
            const icon = WikiDOM.element('div', 'file-icon');
            icon.appendChild(WikiDOM.icon('fa fa-file-zip-o'));
            const details = WikiDOM.element('div', 'file-details');
            details.style.cssText = 'display: flex; flex-direction: column; overflow: hidden;';
            const name = WikiDOM.element('span', 'file-name', backup.name);
            name.title = String(backup.name ?? '');
            const metadata = WikiDOM.element('span', 'file-meta', `${backup.date ?? ''} • ${sizeFormatted}`);
            metadata.style.cssText = 'font-size: 0.85em; color: var(--text-muted);';
            details.append(name, metadata);
            info.append(icon, details);

            const actions = WikiDOM.element('div', 'file-actions');
            const download = WikiDOM.element('a', 'download-file-btn');
            download.href = WikiDOM.localURL(backup.url);
            download.title = window.i18n ? window.i18n.t('common.download') : 'Download';
            download.download = '';
            download.appendChild(WikiDOM.icon('fa fa-download'));
            const remove = WikiDOM.element('button', 'delete-file-btn');
            remove.type = 'button';
            remove.dataset.filename = String(backup.name ?? '');
            remove.title = window.i18n ? window.i18n.t('common.delete') : 'Delete';
            remove.appendChild(WikiDOM.icon('fa fa-trash'));
            actions.append(download, remove);
            item.append(info, actions);
            
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
