(function() {
    'use strict';

    // Apply the persisted state before stylesheets are loaded so the panel is
    // painted in its final position. Missing, invalid, or unavailable storage
    // fails closed to the unobtrusive retracted state.
    let isRetracted = true;
    try {
        isRetracted = window.sessionStorage.getItem('chapter-links-retracted') !== 'false';
    } catch (error) {
        // Storage may be unavailable because of browser privacy settings.
    }

    document.documentElement.classList.toggle('chapter-links-retracted', isRetracted);
})();
