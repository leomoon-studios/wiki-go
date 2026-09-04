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

    // Chapter-links panel retract/expand toggle
    const chapterLinksPanel = document.querySelector('.chapter-links');
    const chapterLinksToggle = document.querySelector('.chapter-links-toggle');

    if (chapterLinksPanel && chapterLinksToggle) {
        const chapterLinksBody = chapterLinksPanel.querySelector('.chapter-links-body');
        const toggleText = chapterLinksToggle.querySelector('.chapter-links-toggle-text');

        function updateChapterLinksState(forceRetracted) {
            const isRetracted = typeof forceRetracted === 'boolean'
                ? forceRetracted
                : !chapterLinksPanel.classList.contains('retracted');

            chapterLinksPanel.classList.toggle('retracted', isRetracted);
            document.documentElement.classList.toggle('chapter-links-retracted', isRetracted);

            chapterLinksToggle.setAttribute('aria-expanded', String(!isRetracted));
            chapterLinksToggle.setAttribute('aria-label', isRetracted ? 'Expand chapter links' : 'Retract chapter links');

            if (toggleText) {
                toggleText.textContent = isRetracted ? '<<' : '>>';
            }

            if (chapterLinksBody) {
                chapterLinksBody.setAttribute('aria-hidden', String(isRetracted));
            }

            // Persist preference
            try {
                sessionStorage.setItem('chapter-links-retracted', String(isRetracted));
            } catch (e) {
                // Ignore storage errors
            }

            return isRetracted;
        }

        // Set when a drag just ended, so the trailing click/tap that fires on
        // pointer-up does not toggle the panel. Checked inside the toggle handler
        // because listener registration order cannot be relied upon here.
        let suppressNextClick = false;

        chapterLinksToggle.addEventListener('click', function() {
            if (suppressNextClick) {
                suppressNextClick = false;
                return;
            }
            updateChapterLinksState(!chapterLinksPanel.classList.contains('retracted'));
        });

        // Touch devices: allow users to drag the toggle tab vertically so it
        // can be moved out of the way of important content.
        (function initMobileDrag() {
            const isCoarse = window.matchMedia('(pointer: coarse)').matches;
            if (!isCoarse) return;

            const TAB_HEIGHT = chapterLinksToggle.offsetHeight || 40;
            const minTop = 60; // keep below the breadcrumbs/header area
            const maxTop = Math.max(minTop, window.innerHeight - TAB_HEIGHT - 16);

            function clampTop(value) {
                return Math.min(Math.max(value, minTop), maxTop);
            }

            // Restore a previously saved vertical offset.
            try {
                const savedOffset = sessionStorage.getItem('chapter-links-toggle-top');
                if (savedOffset !== null) {
                    chapterLinksToggle.style.top = clampTop(parseInt(savedOffset, 10)) + 'px';
                }
            } catch (e) {
                // Ignore storage errors
            }

            let dragging = false;
            let startY = 0;
            let startTop = 0;
            let moved = false;

            chapterLinksToggle.style.touchAction = 'none';

            function onPointerDown(e) {
                // Only drag on touch input.
                if (e.pointerType !== 'touch') return;
                dragging = true;
                moved = false;
                suppressNextClick = false;
                startY = e.clientY;
                startTop = chapterLinksToggle.getBoundingClientRect().top;
                chapterLinksToggle.setPointerCapture(e.pointerId);
                chapterLinksToggle.style.transition = 'none';
            }

            function onPointerMove(e) {
                if (!dragging) return;
                const deltaY = e.clientY - startY;
                if (Math.abs(deltaY) > 3) {
                    moved = true;
                }
                chapterLinksToggle.style.top = clampTop(startTop + deltaY) + 'px';
            }

            function onPointerUp(e) {
                if (!dragging) return;
                dragging = false;
                chapterLinksToggle.style.transition = '';
                try {
                    if (moved) {
                        const top = parseInt(chapterLinksToggle.style.top, 10);
                        sessionStorage.setItem('chapter-links-toggle-top', String(top));
                        suppressNextClick = true;
                    }
                } catch (err) {
                    // Ignore storage errors
                }
            }

            chapterLinksToggle.addEventListener('pointerdown', onPointerDown);
            chapterLinksToggle.addEventListener('pointermove', onPointerMove);
            chapterLinksToggle.addEventListener('pointerup', onPointerUp);
            chapterLinksToggle.addEventListener('pointercancel', onPointerUp);
        })();

        // Restore preference on page load. New sessions default to retracted.
        // Use the two-step transition trick to ensure the browser actually animates
        // the retraction when the page first renders.
        try {
            const saved = sessionStorage.getItem('chapter-links-retracted');
            const savedRetracted = saved === null || saved === 'true';
            if (savedRetracted && !chapterLinksPanel.classList.contains('retracted')) {
                // Set the retracted state without animation first, then enable
                // the transition class so subsequent toggles animate smoothly.
                chapterLinksPanel.classList.add('no-animate');
                updateChapterLinksState(true);
                requestAnimationFrame(function() {
                    requestAnimationFrame(function() {
                        chapterLinksPanel.classList.remove('no-animate');
                    });
                });
            }
        } catch (e) {
            // Ignore storage errors
        }
    }
});