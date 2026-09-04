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
        const expandLabel = chapterLinksToggle.dataset.labelExpand;
        const retractLabel = chapterLinksToggle.dataset.labelRetract;

        function updateChapterLinksState(forceRetracted, persistPreference = true) {
            const isRetracted = typeof forceRetracted === 'boolean'
                ? forceRetracted
                : !chapterLinksPanel.classList.contains('retracted');

            chapterLinksPanel.classList.toggle('retracted', isRetracted);
            document.documentElement.classList.toggle('chapter-links-retracted', isRetracted);

            chapterLinksToggle.setAttribute('aria-expanded', String(!isRetracted));
            chapterLinksToggle.setAttribute('aria-label', isRetracted ? expandLabel : retractLabel);
            chapterLinksPanel.setAttribute('aria-hidden', String(isRetracted));
            if (isRetracted && chapterLinksPanel.contains(document.activeElement)) {
                chapterLinksToggle.focus();
            }
            chapterLinksPanel.toggleAttribute('inert', isRetracted);

            if (chapterLinksBody) {
                chapterLinksBody.setAttribute('aria-hidden', String(isRetracted));
            }

            if (persistPreference) {
                try {
                    sessionStorage.setItem('chapter-links-retracted', String(isRetracted));
                } catch (e) {
                    // Ignore storage errors; the in-page control still works.
                }
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

        // On narrow layouts the panel covers much of the page. Treat a tap on
        // the uncovered page area as a dismissal, while leaving interactions
        // with the panel and its toggle untouched.
        const mobileChapterLinks = window.matchMedia('(max-width: 950px)');
        document.addEventListener('pointerdown', function(e) {
            if (!mobileChapterLinks.matches || chapterLinksPanel.classList.contains('retracted')) {
                return;
            }
            if (chapterLinksPanel.contains(e.target) || chapterLinksToggle.contains(e.target)) {
                return;
            }
            updateChapterLinksState(true);
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
                if (moved) {
                    // Suppress the synthetic trailing click even if storage is
                    // unavailable; persistence must not affect interaction.
                    suppressNextClick = true;
                    const top = parseInt(chapterLinksToggle.style.top, 10);
                    try {
                        sessionStorage.setItem('chapter-links-toggle-top', String(top));
                    } catch (err) {
                        // Ignore storage errors
                    }
                }
            }

            chapterLinksToggle.addEventListener('pointerdown', onPointerDown);
            chapterLinksToggle.addEventListener('pointermove', onPointerMove);
            chapterLinksToggle.addEventListener('pointerup', onPointerUp);
            chapterLinksToggle.addEventListener('pointercancel', onPointerUp);
        })();

        // The head startup script applies this root class before first paint.
        // Mirror it onto the panel without rewriting the stored preference, then
        // enable transitions for subsequent user actions.
        const initiallyRetracted = document.documentElement.classList.contains('chapter-links-retracted');
        chapterLinksPanel.classList.add('no-animate');
        updateChapterLinksState(initiallyRetracted, false);
        requestAnimationFrame(function() {
            requestAnimationFrame(function() {
                chapterLinksPanel.classList.remove('no-animate');
            });
        });
    }
});
