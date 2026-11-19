/**
 * Theme Manager Module
 * Handles theme switching, initialization, and UI updates
 */

// Global print state - set this early
window.isPrintActive = false;

// Track print events globally and early
window.addEventListener('beforeprint', function() {
    window.isPrintActive = true;
});

window.addEventListener('afterprint', function() {
    setTimeout(() => {
        window.isPrintActive = false;
    }, 100);
});

// Immediately set theme before page renders to prevent flash
(function() {
    var savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
        document.documentElement.setAttribute('data-theme', savedTheme);
    } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        document.documentElement.setAttribute('data-theme', 'dark');
    }
})();

document.addEventListener('DOMContentLoaded', function() {
    'use strict';

    // DOM elements
    const themeToggle = document.querySelector('.sidebar-footer-buttons .sidebar-footer-btn[aria-label="Toggle theme"]');
    const root = document.documentElement;
    const lightIcon = document.querySelector('.light-icon');
    const darkIcon = document.querySelector('.dark-icon');

    // Initialize theme
    initializeTheme();

    // Add theme toggle event listener
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }

    // Apply system preference changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', handleSystemThemeChange);

    /**
     * Initialize theme based on saved preference or system preference
     */
    function initializeTheme() {
        const savedTheme = localStorage.getItem('theme');
        const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

        if (savedTheme) {
            applyTheme(savedTheme);
        } else if (systemPrefersDark) {
            applyTheme('dark');
        } else {
            applyTheme('light');
        }
    }

    /**
     * Toggle between light and dark themes
     */
    function toggleTheme() {
        // Don't change theme during print
        if (window.isPrintActive) {
            return;
        }

        const currentTheme = root.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

        applyTheme(newTheme);
        localStorage.setItem('theme', newTheme);
    }

    /**
     * Apply theme to document and update UI
     * @param {string} theme - 'light' or 'dark'
     */
    function applyTheme(theme) {
        // Don't apply theme changes during print
        if (window.isPrintActive) {
            return;
        }

        root.setAttribute('data-theme', theme);
        updateThemeUI(theme);
    }

    /**
     * Update UI elements based on current theme
     * @param {string} theme - 'light' or 'dark'
     */
    function updateThemeUI(theme) {
        if (theme === 'dark') {
            if (lightIcon) lightIcon.style.display = 'none';
            if (darkIcon) darkIcon.style.display = 'block';
            // Use dark theme for syntax highlighting
            const prismTheme = document.getElementById('prism-theme');
            if (prismTheme) {
                prismTheme.href = wikiURL('/static/libs/prism-1.30.0/prism-tomorrow.min.css');
            }
        } else {
            if (lightIcon) lightIcon.style.display = 'block';
            if (darkIcon) darkIcon.style.display = 'none';
            // Use light theme for syntax highlighting
            const prismTheme = document.getElementById('prism-theme');
            if (prismTheme) {
                prismTheme.href = wikiURL('/static/libs/prism-1.30.0/prism.min.css');
            }
        }

        // Dispatch theme change event for other components
        const themeChangeEvent = new CustomEvent('themeChanged', {
            detail: { theme: theme }
        });
        document.dispatchEvent(themeChangeEvent);
    }

    /**
     * Handle system theme preference changes
     * @param {MediaQueryListEvent} e - Media query change event
     */
    function handleSystemThemeChange(e) {
        // Don't change theme during print or if user has set preference
        if (window.isPrintActive || localStorage.getItem('theme')) {
            return;
        }
        const newTheme = e.matches ? 'dark' : 'light';
        applyTheme(newTheme);
    }

    // Export theme functions to global scope
    window.ThemeManager = {
        getCurrentTheme: function() {
            return root.getAttribute('data-theme') || 'light';
        },
        setTheme: applyTheme,
        toggleTheme: toggleTheme,
        updateUI: updateThemeUI
    };
});