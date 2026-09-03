// Dialog System Module for Wiki-Go
// Provides message and confirmation dialog functionality
(function() {
    'use strict';

    // Module state
    let messageDialog;
    let messageTitle;
    let messageContent;
    let messageOkButton;
    let closeMessageDialog;
    let messageCallback = null;
    
    let userConfirmDialog;
    let confirmTitle;
    let confirmContent;
    let confirmYesButton;
    let confirmNoButton;
    let closeConfirmDialog;
    
    // Initialize early references for immediate use before DOM is loaded
    // These will be replaced by the fully initialized versions once the DOM is loaded
    window.showConfirmDialog = function(title, message, callback) {
        document.addEventListener('DOMContentLoaded', () => {
            const userConfirmDialog = document.querySelector('.user-confirmation-dialog');
            const confirmTitle = document.querySelector('.user-confirmation-dialog .confirm-title');
            const confirmContent = document.querySelector('.user-confirmation-dialog .confirm-content');

            confirmTitle.textContent = title;
            confirmContent.textContent = message;

            // Store callback for the buttons to use
            window.confirmCallback = callback;
            userConfirmDialog.classList.add('active');
        });
    };

    window.showMessageDialog = function(title, message, callback) {
        document.addEventListener('DOMContentLoaded', () => {
            const messageDialog = document.querySelector('.message-dialog');
            const messageTitle = document.querySelector('.message-dialog .message-title');
            const messageContent = document.querySelector('.message-dialog .message-content');

            messageTitle.textContent = title;
            messageContent.textContent = message;
            messageCallback = typeof callback === 'function' ? callback : null;
            messageDialog.classList.add('active');
        });
    };

    // Initialize dialog elements when DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        // Message dialog elements
        messageDialog = document.querySelector('.message-dialog');
        messageTitle = document.querySelector('.message-dialog .message-title');
        messageContent = document.querySelector('.message-dialog .message-content');
        messageOkButton = document.querySelector('.message-dialog .message-ok');
        closeMessageDialog = messageDialog.querySelector('.close-dialog');

        // Confirmation dialog elements
        userConfirmDialog = document.querySelector('.user-confirmation-dialog');
        confirmTitle = document.querySelector('.user-confirmation-dialog .confirm-title');
        confirmContent = document.querySelector('.user-confirmation-dialog .confirm-content');
        confirmYesButton = document.querySelector('.user-confirmation-dialog .confirm-yes');
        confirmNoButton = document.querySelector('.user-confirmation-dialog .confirm-no');
        closeConfirmDialog = userConfirmDialog.querySelector('.close-dialog');
        
        // Set up event listeners for the message dialog
        if (messageOkButton) {
            messageOkButton.addEventListener('click', hideMessageDialog);
        }

        if (closeMessageDialog) {
            closeMessageDialog.addEventListener('click', hideMessageDialog);
        }
        
        // Set up event listeners for the confirmation dialog
        if (confirmYesButton) {
            confirmYesButton.addEventListener('click', function() {
                hideConfirmDialog();
                if (window.confirmCallback) {
                    window.confirmCallback(true);
                    window.confirmCallback = null; // Clear callback after use
                }
            });
        }

        if (confirmNoButton) {
            confirmNoButton.addEventListener('click', function() {
                hideConfirmDialog();
                if (window.confirmCallback) {
                    window.confirmCallback(false);
                    window.confirmCallback = null; // Clear callback after use
                }
            });
        }

        if (closeConfirmDialog) {
            closeConfirmDialog.addEventListener('click', function() {
                hideConfirmDialog();
                if (window.confirmCallback) {
                    window.confirmCallback(false);
                    window.confirmCallback = null; // Clear callback after use
                }
            });
        }
        
        // Replace the early references with the proper implementations
        window.showMessageDialog = showMessageDialog;
        window.showConfirmDialog = showConfirmDialog;
        window.hideMessageDialog = hideMessageDialog;
        window.hideConfirmDialog = hideConfirmDialog;
    });
    
    // Message dialog functions
    function showMessageDialog(title, message, callback) {
        if (!messageTitle || !messageContent || !messageDialog) {
            console.error('Message dialog elements not initialized');
            return;
        }
        
        messageTitle.textContent = title;
        messageContent.textContent = message;
        messageCallback = typeof callback === 'function' ? callback : null;
        messageDialog.classList.add('active');
    }

    function hideMessageDialog() {
        if (!messageDialog) {
            console.error('Message dialog element not initialized');
            return;
        }
        
        messageDialog.classList.remove('active');
        const callback = messageCallback;
        messageCallback = null;
        if (callback) {
            callback();
        }
    }
    
    // Confirmation dialog functions
    function showConfirmDialog(title, message, callback) {
        if (!confirmTitle || !confirmContent || !userConfirmDialog) {
            console.error('Confirmation dialog elements not initialized');
            return;
        }
        
        confirmTitle.textContent = title;
        confirmContent.textContent = message;
        window.confirmCallback = callback; // Store callback globally for buttons to use
        userConfirmDialog.classList.add('active');
    }

    function hideConfirmDialog() {
        if (!userConfirmDialog) {
            console.error('Confirmation dialog element not initialized');
            return;
        }
        
        userConfirmDialog.classList.remove('active');
    }

    // Expose dialog functions globally
    window.DialogSystem = {
        showMessageDialog: showMessageDialog,
        hideMessageDialog: hideMessageDialog,
        showConfirmDialog: showConfirmDialog,
        hideConfirmDialog: hideConfirmDialog
    };
})();
