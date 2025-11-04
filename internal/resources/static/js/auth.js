// Authentication Module for Wiki-Go
// Handles login, logout, and authentication-related functionality
(function() {
    'use strict';

    // Module state
    let loginDialog;
    let closeDialog;
    let loginForm;
    let errorMessage;
    let loginUsernameInput;
    let editCallback = null;

    // Initialize module when DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        // Login Dialog Elements
        loginDialog = document.querySelector('.login-dialog');
        closeDialog = document.querySelector('.login-dialog .close-dialog');
        loginForm = document.getElementById('loginForm');
        errorMessage = document.querySelector('.login-dialog .error-message');
        loginUsernameInput = document.getElementById('username');

        // Set up event listeners
        if (closeDialog) {
            closeDialog.addEventListener('click', hideLoginDialog);
        }

        if (loginForm) {
            loginForm.addEventListener('submit', handleLoginSubmit);
        }

        // Add click handler for login button
        const loginButton = document.querySelector('.toolbar-button.auth-button.primary');
        if (loginButton) {
            loginButton.addEventListener('click', function() {
                // Show the login dialog
                loginDialog.classList.add('active');

                // Reset form and clear error messages
                loginForm.reset();
                errorMessage.style.display = 'none';

                // Focus on username field
                setTimeout(() => {
                    loginUsernameInput.focus();
                }, 100);

                // Set callback for after successful login
                window.loginCallback = function() {
                    updateToolbarButtons();
                };
            });
        }

        // Add click handler for logout button
        const logoutButton = document.querySelector('.logout-button');
        if (logoutButton) {
            logoutButton.addEventListener('click', handleLogout);
        }

        // Escape key is now handled by keyboard-shortcuts.js

        // Add click handler for login dialog close button
        document.addEventListener('click', function(e) {
            if (e.target.closest('.login-dialog .close-dialog')) {
                hideLoginDialog();
            }
        });

        // Check if default password is in use
        checkDefaultPassword();

        // Check for pending actions from previous login - remove the delay
        checkPendingActions();
    });

    // Function to show login dialog
    function showLoginDialog(callback) {
        loginDialog.classList.add('active');
        editCallback = callback;
        errorMessage.style.display = 'none';
        loginForm.reset();
        // Focus on username field after dialog is shown
        setTimeout(() => {
            loginUsernameInput.focus();
        }, 100);
    }

    // Function to hide login dialog
    function hideLoginDialog() {
        if (loginDialog) {
            loginDialog.classList.remove('active');
        }
    }

    // Function to handle login form submission
    async function handleLoginSubmit(e) {
        e.preventDefault();
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const keepLoggedIn = document.getElementById('keepLoggedIn')?.checked || false;

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    username,
                    password,
                    keepLoggedIn
                })
            });

            if (response.ok) {
                hideLoginDialog();
                if (window.loginCallback) {
                    // Store loginCallback info in localStorage
                    localStorage.setItem('pendingAction', 'loginCallback');
                    window.loginCallback = null; // Clear the callback after use
                } else if (editCallback) {
                    // Store edit action in localStorage
                    localStorage.setItem('pendingAction', 'editPage');
                }

                // Reload the page to refresh the comments section
                window.location.reload();
            } else {
                let msg = window.i18n ? window.i18n.t('login.error') : 'Invalid username or password';

                if (response.status === 429) {
                    try {
                        const data = await response.json();
                        if (data && data.message) {
                            msg = data.message;
                            if (data.retryAfter) {
                                const retryTxt = window.i18n ? window.i18n.t('login.retry_in') : 'retry in';
                                msg += ` (${retryTxt} ${data.retryAfter}s)`;
                            }
                        }
                    } catch (e) {}
                }

                errorMessage.textContent = msg;
                errorMessage.style.display = 'block';
            }
        } catch (error) {
            console.error('Login error:', error);
            errorMessage.textContent = 'An error occurred. Please try again.';
            errorMessage.style.display = 'block';
        }
    }

    // Function to handle logout
    async function handleLogout() {
        try {
            const response = await fetch('/api/logout', {
                method: 'POST'
            });

            if (response.ok) {
                // Update toolbar buttons after logout
                updateToolbarButtons();

                // Reload the page to refresh the comments section
                window.location.reload();
            } else {
                window.DialogSystem.showMessageDialog('Error', 'Failed to logout');
            }
        } catch (error) {
            console.error('Error during logout:', error);
            window.DialogSystem.showMessageDialog('Error', 'Failed to logout');
        }
    }

    // Function to check if current user is an admin
    async function checkIfUserIsAdmin() {
        try {
            const response = await fetch('/api/check-auth');
            if (!response.ok) {
                return false;
            }

            const data = await response.json();
            return data.role === 'admin';
        } catch (error) {
            console.error('Error checking admin status:', error);
            return false;
        }
    }

    // Function to check if current user has a specific role
    async function checkUserRole(requiredRole) {
        try {
            const response = await fetch('/api/check-auth');
            if (!response.ok) {
                return false;
            }

            const data = await response.json();

            // Handle role hierarchy
            if (requiredRole === 'admin') {
                return data.role === 'admin';
            } else if (requiredRole === 'editor') {
                return data.role === 'admin' || data.role === 'editor';
            } else if (requiredRole === 'viewer') {
                return data.role === 'admin' || data.role === 'editor' || data.role === 'viewer';
            }

            return false;
        } catch (error) {
            console.error('Error checking user role:', error);
            return false;
        }
    }

    // Function to show permission error
    function showPermissionError(requiredRole) {
        let message = "You don't have permission to perform this action.";
        if (requiredRole) {
            message = `This feature requires ${requiredRole} privileges.`;
        }
        window.DialogSystem.showMessageDialog("Permission Denied", message);
    }

    // Function to show admin-only feature error (deprecated, use showPermissionError instead)
    function showAdminOnlyError() {
        window.DialogSystem.showMessageDialog("Admin Access Required", "This feature is only available to administrators.");
    }

    // Function to update toolbar buttons based on authentication status
    async function updateToolbarButtons() {
        try {
            // Check authentication status
            const authResponse = await fetch('/api/check-auth');

            if (authResponse.status === 401) {
                // User is not authenticated
                // Hide all admin-only buttons
                document.querySelectorAll('.admin-only-button').forEach(btn => {
                    btn.style.display = 'none';
                });

                // Show login button, hide logout button
                document.querySelector('.toolbar-button.auth-button.primary').style.cssText = 'display: inline-flex !important';
                document.querySelector('.logout-button').style.cssText = 'display: none !important';
                return;
            }

            // User is authenticated, check role
            const authData = await authResponse.json();
            const isAdmin = authData.role === 'admin';
            const isEditor = authData.role === 'editor';

            // Show/hide buttons based on role
            if (isAdmin) {
                // Admin user - show all admin buttons
                document.querySelectorAll('.admin-only-button').forEach(btn => {
                    btn.style.cssText = 'display: inline-flex !important';
                });

                // Show editor buttons
                document.querySelectorAll('.editor-only-button').forEach(btn => {
                    btn.style.cssText = 'display: inline-flex !important';
                });

                // Special case for move/rename button (only show if not on homepage)
                const renameBtn = document.querySelector('.move-document');
                if (renameBtn && (window.location.pathname === '/' || window.location.pathname === '/homepage')) {
                    renameBtn.style.cssText = 'display: none !important';
                }
            } else if (isEditor) {
                // Editor user - hide admin buttons, show editor buttons
                document.querySelectorAll('.admin-only-button').forEach(btn => {
                    btn.style.display = 'none';
                });

                document.querySelectorAll('.editor-only-button').forEach(btn => {
                    btn.style.cssText = 'display: inline-flex !important';
                });

                // Special case for move/rename button (only show if not on homepage)
                const renameBtn = document.querySelector('.move-document');
                if (renameBtn && (window.location.pathname === '/' || window.location.pathname === '/homepage')) {
                    renameBtn.style.cssText = 'display: none !important';
                }
            } else {
                // Viewer user - hide admin and editor buttons
                document.querySelectorAll('.admin-only-button, .editor-only-button').forEach(btn => {
                    btn.style.display = 'none';
                });
            }

            // Show logout button, hide login button for all authenticated users
            document.querySelector('.toolbar-button.auth-button.primary').style.cssText = 'display: none !important';
            document.querySelector('.logout-button').style.cssText = 'display: inline-flex !important';
        } catch (error) {
            console.error('Error checking authentication status:', error);
        }
    }

    // Function to check if the default password is in use
    async function checkDefaultPassword() {
        try {
            const response = await fetch('/api/check-default-password');
            const data = await response.json();

            const banner = document.getElementById('password-warning-banner');
            if (!banner) return;

            if (data.defaultPasswordInUse) {
                banner.style.display = 'block';
                document.body.classList.add('has-password-warning');
            } else {
                banner.style.display = 'none';
                document.body.classList.remove('has-password-warning');
            }
        } catch (error) {
            console.error('Error checking default password:', error);
        }
    }

    // Function to check for pending actions from previous login
    function checkPendingActions() {
        const pendingAction = localStorage.getItem('pendingAction');
        if (pendingAction) {
            // Clear the pending action first to prevent loops
            localStorage.removeItem('pendingAction');

            if (pendingAction === 'editPage') {
                // Check if the user has editor or admin role then navigate to edit mode
                checkUserRole('editor').then(canEdit => {
                    if (canEdit) {
                        // Navigate to edit mode
                        const url = new URL(window.location);
                        url.searchParams.set('mode', 'edit');
                        window.location.href = url.toString();
                    } else {
                        showPermissionError('editor');
                    }
                });
            } else if (pendingAction === 'loginCallback') {
                updateToolbarButtons();
            }
        }
    }

    // Expose public API
    window.Auth = {
        showLoginDialog: showLoginDialog,
        hideLoginDialog: hideLoginDialog,
        checkIfUserIsAdmin: checkIfUserIsAdmin,
        checkUserRole: checkUserRole,
        showAdminOnlyError: showAdminOnlyError,
        showPermissionError: showPermissionError,
        updateToolbarButtons: updateToolbarButtons,
        checkDefaultPassword: checkDefaultPassword
    };
})();
