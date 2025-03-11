// Initialize Mermaid diagrams
document.addEventListener('DOMContentLoaded', function() {
    // Simple initialization for Mermaid
    if (typeof mermaid !== 'undefined') {
        // Get current theme
        const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';

        // Initialize with theme-specific settings
        mermaid.initialize({
            startOnLoad: true,
            theme: isDarkMode ? 'dark' : 'default',
            securityLevel: 'strict',
            flowchart: {
                useMaxWidth: true,
                htmlLabels: true
            }
        });

        console.log('Mermaid initialized with theme:', isDarkMode ? 'dark' : 'default');
    } else {
        console.error('Mermaid library not loaded');
    }

    // Add theme change listener
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', function() {
            // Wait for theme to change
            setTimeout(function() {
                if (typeof mermaid !== 'undefined') {
                    try {
                        // Get updated theme
                        const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';
                        console.log('Theme changed to:', isDarkMode ? 'dark' : 'light');

                        // Re-initialize with new theme
                        mermaid.initialize({
                            startOnLoad: false,
                            theme: isDarkMode ? 'dark' : 'default'
                        });

                        // Re-render all diagrams
                        document.querySelectorAll('.mermaid').forEach(function(el) {
                            // Store original content if not already stored
                            if (!el.getAttribute('data-source')) {
                                el.setAttribute('data-source', el.textContent);
                            }

                            const source = el.getAttribute('data-source');
                            const id = 'mermaid-' + Math.random().toString(36).substring(2, 10);

                            // Clear and re-render
                            el.innerHTML = '';
                            mermaid.render(id, source, function(svgCode) {
                                el.innerHTML = svgCode;
                            }, el);
                        });
                    } catch (e) {
                        console.error('Error re-rendering Mermaid diagrams:', e);
                    }
                }
            }, 100);
        });
    }
});