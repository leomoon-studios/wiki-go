// Login page behavior. Server-provided messages are read from escaped data attributes.
(function() {
    'use strict';

    document.addEventListener('DOMContentLoaded', function() {
        const loginForm = document.getElementById('loginForm');
        const errorMessage = document.getElementById('loginError');
        if (!loginForm || !errorMessage) return;

        const errorText = errorMessage.dataset.errorMessage;

        loginForm.addEventListener('submit', async function(event) {
            event.preventDefault();

            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const keepLoggedIn = document.getElementById('keepLoggedIn').checked;
            const submitButton = loginForm.querySelector('button[type="submit"]');
            const originalText = submitButton.textContent;

            submitButton.disabled = true;
            submitButton.textContent = 'Logging in...';

            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username,
                        password,
                        keepLoggedIn
                    })
                });

                if (response.ok) {
                    const params = new URLSearchParams(window.location.search);
                    const destination = params.get('redirect');
                    window.location.href = destination && destination.startsWith('/') ? destination : '/';
                    return;
                }

                let message = errorText;
                if (response.status === 429) {
                    try {
                        const data = await response.json();
                        if (data && data.message) {
                            message = data.message;
                            if (data.retryAfter) {
                                const retryText = window.i18n ? window.i18n.t('login.retry_in') : 'retry in';
                                message += ` (${retryText} ${data.retryAfter}s)`;
                            }
                        }
                    } catch (_error) {
                        // Keep the translated generic login error.
                    }
                }

                errorMessage.textContent = message;
                errorMessage.style.display = 'block';
                submitButton.disabled = false;
                submitButton.textContent = originalText;
            } catch (error) {
                console.error('Login error:', error);
                errorMessage.textContent = 'An error occurred. Please try again.';
                errorMessage.style.display = 'block';
                submitButton.disabled = false;
                submitButton.textContent = originalText;
            }
        });
    });
})();
