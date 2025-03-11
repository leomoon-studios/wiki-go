/**
 * Markdown Extensions JavaScript
 * Handles interactive behavior for custom markdown extensions
 */

// Handle collapsible sections in print view
document.addEventListener('DOMContentLoaded', function() {
    // Find all collapsible sections
    const collapsibleSections = document.querySelectorAll('details.markdown-details');

    if (collapsibleSections.length > 0) {
        // Store original states
        const originalStates = new Map();

        // Add event listeners for print
        window.addEventListener('beforeprint', function() {
            // Store original states and open all sections before printing
            collapsibleSections.forEach(details => {
                originalStates.set(details, details.open);
                details.open = true;
            });
        });

        window.addEventListener('afterprint', function() {
            // Restore original states after printing
            collapsibleSections.forEach(details => {
                if (originalStates.has(details)) {
                    details.open = originalStates.get(details);
                }
            });
        });
    }
});