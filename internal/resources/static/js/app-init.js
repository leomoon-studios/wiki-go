// Application Initialization Module

// Apply content width setting immediately before DOM content is loaded to prevent flashing
(function() {
    const disableMaxWidth = document.querySelector('meta[name="disable-content-max-width"]')?.getAttribute('content') === 'true';
    if (disableMaxWidth) {
        // Apply a style immediately to prevent flashing
        document.documentElement.style.setProperty('--content-max-width', 'none');
    }
})();

document.addEventListener('DOMContentLoaded', function() {
    'use strict';

    // Initialize Wiki configuration
    window.WikiConfig = {
        // Get config values from meta tags
        enableLinkEmbedding: document.querySelector('meta[name="enable-link-embedding"]')?.getAttribute('content') === 'true',
        disableContentMaxWidth: document.querySelector('meta[name="disable-content-max-width"]')?.getAttribute('content') === 'true'
    };

    // Apply full width content if setting is enabled
    if (window.WikiConfig.disableContentMaxWidth) {
        document.querySelector('.content')?.classList.add('full-width-content');
        document.querySelector('.search-results')?.classList.add('full-width-content');
    }

    // Update content width when settings change
    // This will be triggered after settings are updated
    document.addEventListener('settings-updated', function() {
        const contentElement = document.querySelector('.content');
        const searchResultsElement = document.querySelector('.search-results');
        const disableMaxWidth = document.querySelector('meta[name="disable-content-max-width"]')?.getAttribute('content') === 'true';

        if (disableMaxWidth) {
            contentElement?.classList.add('full-width-content');
            searchResultsElement?.classList.add('full-width-content');
            document.documentElement.style.setProperty('--content-max-width', 'none');
        } else {
            contentElement?.classList.remove('full-width-content');
            searchResultsElement?.classList.remove('full-width-content');
            document.documentElement.style.removeProperty('--content-max-width');
        }
    });

    // Initialize editor controls
    if (window.WikiEditor && typeof window.WikiEditor.initializeEditControls === 'function') {
        window.WikiEditor.initializeEditControls();
    }

    // Add scroll event listener to toggle shadows
    const breadcrumbs = document.querySelector('.breadcrumbs');
    const hamburger = document.querySelector('.hamburger');

    // Function to check scroll position and update shadows
    function updateShadows() {
        if (window.scrollY > 10) {
            if (breadcrumbs) breadcrumbs.classList.add('scrolled');
            if (hamburger) hamburger.classList.add('scrolled');
        } else {
            if (breadcrumbs) breadcrumbs.classList.remove('scrolled');
            if (hamburger) hamburger.classList.remove('scrolled');
        }
    }

    // Check on page load
    updateShadows();

    // Check on scroll
    window.addEventListener('scroll', updateShadows);

    // Dialog scroll shadows
    const dialogContainers = document.querySelectorAll('.dialog-container, .login-container');
    dialogContainers.forEach(container => {
        // Find all potential scrollable elements within the container
        // We use :scope > for direct children where applicable, but also look deeper for version history
        const directScrollables = Array.from(container.querySelectorAll(':scope > form, :scope > .tab-content, :scope > .dialog-message, :scope > .message-content, :scope > .confirm-content, :scope > .import-management, :scope > .access-rule-form, :scope > .new-document-form, :scope > .settings-form, :scope > .login-form'));
        
        // Version history has nested scrollable areas (desktop) or the main content (mobile)
        // We also check .dialog-content which is scrollable in mobile version history
        const versionScrollables = Array.from(container.querySelectorAll('.version-list-container, .version-preview-container, .dialog-content'));
        
        const allScrollables = [...directScrollables, ...versionScrollables];
        
        // Fallback for older browsers or if structure is different (simple search)
        if (allScrollables.length === 0) {
             const fallback = container.querySelector('.tab-content, form, .dialog-message, .message-content, .confirm-content');
             if (fallback) allScrollables.push(fallback);
        }
        
        if (allScrollables.length > 0) {
            const checkScroll = () => {
                // If ANY of the scrollable areas are scrolled, show the shadow
                const isScrolled = allScrollables.some(el => el.scrollTop > 0);
                if (isScrolled) {
                    container.classList.add('scrolled');
                } else {
                    container.classList.remove('scrolled');
                }
            };
            
            allScrollables.forEach(el => {
                el.addEventListener('scroll', checkScroll);
            });
        }
    });
});
