// Comments functionality
document.addEventListener('DOMContentLoaded', function() {
    // Translate comments section after DOM loads
    setTimeout(function() {
        if (window.i18n) {
            // Translate manually for comments section
            document.querySelectorAll('.comments-section h3').forEach(el => {
                if (el.textContent === 'comments.title') {
                    el.textContent = window.i18n.t('comments.title');
                }
            });

            document.querySelectorAll('.comment-form textarea').forEach(el => {
                if (el.placeholder === 'comments.write_placeholder') {
                    el.placeholder = window.i18n.t('comments.write_placeholder');
                }
            });

            document.querySelectorAll('.comment-form button').forEach(el => {
                if (el.textContent === 'comments.post_button') {
                    el.textContent = window.i18n.t('comments.post_button');
                }
            });

            document.querySelectorAll('.comment-form .form-help').forEach(el => {
                if (el.textContent === 'comments.markdown_supported') {
                    el.textContent = window.i18n.t('comments.markdown_supported');
                }
            });

            document.querySelectorAll('.login-prompt').forEach(el => {
                if (el.textContent.includes('comments.login_required')) {
                    const loginLink = WikiDOM.element('a', 'open-login', window.i18n.t('comments.login'));
                    loginLink.href = '#';
                    loginLink.addEventListener('click', function(e) {
                        e.preventDefault();
                        showLoginDialog();
                    });
                    el.replaceChildren(
                        document.createTextNode(window.i18n.t('comments.login_required') + ' '),
                        loginLink
                    );
                }
            });

            document.querySelectorAll('.no-comments').forEach(el => {
                if (el.textContent === 'comments.no_comments') {
                    el.textContent = window.i18n.t('comments.no_comments');
                }
            });

            document.querySelectorAll('.delete-comment').forEach(el => {
                if (el.title === 'comments.delete_title') {
                    el.title = window.i18n.t('comments.delete_title');
                }
            });
        }
    }, 500); // Small delay to ensure i18n is loaded

    // Handle comment form submission
    const commentForm = document.getElementById('comment-form');
    if (commentForm) {
        commentForm.addEventListener('submit', async function(e) {
            e.preventDefault();

            // Get the comment content
            const commentContent = this.querySelector('textarea[name="content"]').value;
            if (!commentContent.trim()) {
                return; // Don't submit empty comments
            }

            // Get the current document path
            const docPath = getCurrentDocPath();

            try {
                // Show loading state
                const submitButton = this.querySelector('button[type="submit"]');
                const originalText = submitButton.textContent;
                submitButton.textContent = window.i18n ? window.i18n.t('common.sending') : 'Sending...';
                submitButton.disabled = true;

                // Send the comment to the server
                const response = await fetch(`/api/comments/add/${docPath}`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ content: commentContent })
                });

                // Check if the request was successful
                if (response.ok) {
                    // Clear the form
                    this.reset();

                    // Reload the page to show the new comment
                    window.location.reload();
                } else {
                    // Show error message
                    const data = await response.json();
                    showMessageDialog(
                        window.i18n ? window.i18n.t('comments.error_title') : 'Error',
                        data.message || (window.i18n ? window.i18n.t('comments.error_generic') : 'Failed to post comment')
                    );
                }
            } catch (error) {
                // Show error message
                showMessageDialog(
                    window.i18n ? window.i18n.t('comments.error_title') : 'Error',
                    window.i18n ? window.i18n.t('comments.error_generic') : 'Failed to post comment'
                );
                console.error('Error posting comment:', error);
            } finally {
                // Restore button state
                submitButton.textContent = originalText;
                submitButton.disabled = false;
            }
        });
    }

    // Handle comment deletion (admin only)
    document.querySelectorAll('.delete-comment').forEach(button => {
        button.addEventListener('click', function() {
            const commentId = this.dataset.id;
            const docPath = getCurrentDocPath();

            console.log('Delete button clicked for comment ID:', commentId);
            console.log('Current document path:', docPath);

            // Show confirmation dialog using global function
            console.log('Showing confirmation dialog...');
            if (typeof window.showConfirmDialog === 'function') {
                window.showConfirmDialog(
                    window.i18n ? window.i18n.t('comments.delete_title') : 'Delete Comment',
                    window.i18n ? window.i18n.t('comments.delete_confirm') : 'Are you sure you want to delete this comment? This action cannot be undone.',
                    async (confirmed) => {
                        console.log('User response to delete confirmation:', confirmed);
                        if (!confirmed) return;

                        try {
                            console.log('Sending delete request to:', `/api/comments/delete/${docPath}/${commentId}`);
                            // Send delete request to the server
                            const response = await fetch(`/api/comments/delete/${docPath}/${commentId}`, {
                                method: 'DELETE'
                            });

                            console.log('Delete response status:', response.status);

                            // Check if the request was successful
                            if (response.ok) {
                                console.log('Delete successful, removing comment from page');
                                // Remove the comment from the page
                                const comment = Array.from(document.querySelectorAll('.user-comment'))
                                    .find(element => element.dataset.id === String(commentId));
                                if (comment) {
                                    comment.remove();

                                    // If there are no more comments, show the "no comments" message
                                    const commentsList = document.querySelector('.comments-list');
                                    if (commentsList && !commentsList.querySelector('.user-comment')) {
                                        const noComments = document.createElement('p');
                                        noComments.className = 'no-comments';
                                        noComments.textContent = window.i18n ? window.i18n.t('comments.no_comments') : 'No comments yet.';
                                        commentsList.appendChild(noComments);
                                    }
                                }
                            } else {
                                console.error('Delete failed with status:', response.status);
                                // Show error message
                                const data = await response.json();
                                console.error('Error data:', data);
                                if (typeof window.showMessageDialog === 'function') {
                                    window.showMessageDialog(
                                        window.i18n ? window.i18n.t('comments.error_title') : 'Error',
                                        data.message || (window.i18n ? window.i18n.t('comments.error_delete') : 'Failed to delete comment')
                                    );
                                } else {
                                    alert(data.message || (window.i18n ? window.i18n.t('comments.error_delete') : 'Failed to delete comment'));
                                }
                            }
                        } catch (error) {
                            console.error('Exception during delete operation:', error);
                            // Show error message
                            if (typeof window.showMessageDialog === 'function') {
                                window.showMessageDialog(
                                    window.i18n ? window.i18n.t('comments.error_title') : 'Error',
                                    window.i18n ? window.i18n.t('comments.error_delete') : 'Failed to delete comment'
                                );
                            } else {
                                alert(window.i18n ? window.i18n.t('comments.error_delete') : 'Failed to delete comment');
                            }
                            console.error('Error deleting comment:', error);
                        }
                    }
                );
            } else {
                // Fallback if showConfirmDialog is not available
                console.error('showConfirmDialog function not found, using alert instead');
                if (confirm(window.i18n ? window.i18n.t('comments.delete_confirm') : 'Are you sure you want to delete this comment? This action cannot be undone.')) {
                    // Similar delete logic as above
                    fetch(`/api/comments/delete/${docPath}/${commentId}`, {
                        method: 'DELETE'
                    })
                    .then(response => {
                        if (response.ok) {
                            const comment = Array.from(document.querySelectorAll('.user-comment'))
                                .find(element => element.dataset.id === String(commentId));
                            if (comment) {
                                comment.remove();

                                // Check if there are any comments left
                                const commentsList = document.querySelector('.comments-list');
                                if (commentsList && !commentsList.querySelector('.user-comment')) {
                                    const noComments = document.createElement('p');
                                    noComments.className = 'no-comments';
                                    noComments.textContent = window.i18n ? window.i18n.t('comments.no_comments') : 'No comments yet.';
                                    commentsList.appendChild(noComments);
                                }
                            }
                        } else {
                            alert(window.i18n ? window.i18n.t('comments.error_delete') : 'Failed to delete comment');
                        }
                    })
                    .catch(error => {
                        console.error('Error deleting comment:', error);
                        alert(window.i18n ? window.i18n.t('comments.error_delete') : 'Failed to delete comment');
                    });
                }
            }
        });
    });

    // Handle login link in the comments section
    const loginLink = document.querySelector('.login-prompt .open-login');
    if (loginLink) {
        loginLink.addEventListener('click', function(e) {
            e.preventDefault();
            showLoginDialog();
        });
    }

});
