/**
 * Edit Button Handler
 * This script handles the Edit button click in VIEW mode
 * It must be loaded outside of the {{if .IsEditMode}} conditional
 * because the editor scripts are only loaded in edit mode
 */

document.addEventListener('DOMContentLoaded', function() {
    const editPageButton = document.querySelector('.edit-page');
    
    // Only set up handler if we're NOT in edit mode
    const urlParams = new URLSearchParams(window.location.search);
    const isEditMode = urlParams.get('mode') === 'edit';
    
    if (!isEditMode && editPageButton) {
        editPageButton.addEventListener('click', async function() {
            // Check authentication first
            try {
                const authResponse = await fetch('/api/check-auth');
                if (authResponse.status === 401) {
                    // Show login dialog
                    if (window.Auth && window.Auth.showLoginDialog) {
                        window.Auth.showLoginDialog(() => {
                            // The auth.js checkPendingActions will handle navigation after login
                        });
                        // Set pending action so after reload it navigates to edit mode
                        localStorage.setItem('pendingAction', 'editPage');
                    }
                    return;
                }
                
                // Check if user has editor/admin role
                if (window.Auth && window.Auth.checkUserRole) {
                    const canEdit = await window.Auth.checkUserRole('editor');
                    if (!canEdit) {
                        if (window.Auth.showPermissionError) {
                            window.Auth.showPermissionError('editor');
                        }
                        return;
                    }
                }
                
                // Navigate to edit mode
                const url = new URL(window.location);
                url.searchParams.set('mode', 'edit');
                window.location.href = url.toString();
            } catch (error) {
                console.error('Error checking authentication:', error);
            }
        });
    }
});
