// Sidebar Navigation Module for Wiki-Go
// Handles sidebar functionality, hamburger menu toggle, and mobile touch gestures
(function() {
    'use strict';

    // ========== MODULE STATE ==========
    let hamburger, sidebar, content, body;

    // Storage key for persisting sidebar directory states
    const NAV_STATE_STORAGE_KEY = 'wikiGo.sidebarNavState';
    
    // Touch gesture state
    let touchStartX = 0, touchEndX = 0;
    let touchStartY = 0, touchCurrentY = 0;
    let isDragging = false;
    let dragProgress = 0;
    let startTime = 0;
    
    // ========== CONFIGURATION ==========
    const CONFIG = {
        swipeThreshold: 50,        // Minimum distance for swipe
        edgeThreshold: 50,         // Left edge detection zone
        verticalThreshold: 30,     // Max vertical movement for horizontal swipe
        dragFollowThreshold: 10,   // Min distance before drag follow starts
        maxSwipeTime: 300,         // Max time for valid swipe (ms)
        snapThreshold: 0.4,        // 40% drag = snap to open
        autoScrollOffset: 150      // Default scroll offset for active items
    };

    // ========== INITIALIZATION ==========
    document.addEventListener('DOMContentLoaded', function() {
        // Get DOM elements
        hamburger = document.querySelector('.hamburger');
        sidebar = document.querySelector('.sidebar');
        content = document.querySelector('.content');
        body = document.body;

        // Initialize all features
        initHamburgerMenu();
        initClickOutside();
        initSidebarLinks();
        initNavExpandCollapse();
        initSitemapButton();
        hydrateNavState();
        initTouchGestures();
        scrollActiveIntoView();

        // Persist state before the page unloads (handles all navigation paths)
        window.addEventListener('pagehide', saveNavState);
    });

    function initSitemapButton() {
        const sitemapButton = document.querySelector('.open-sitemap');
        if (!sitemapButton) return;

        sitemapButton.addEventListener('click', function() {
            const sitemapURL = sitemapButton.dataset.sitemapUrl;
            if (sitemapURL) {
                window.open(sitemapURL, '_blank');
            }
        });
    }

    // ========== NAV EXPAND/COLLAPSE ==========

    function initNavExpandCollapse() {
        const navItems = document.querySelector('.nav-items');
        if (!navItems) return;

        // Event delegation so it keeps working after refreshSidebar() re-renders the nav
        navItems.addEventListener('click', function(e) {
            const arrow = e.target.closest('button.nav-arrow');
            if (!arrow) return;

            e.preventDefault();
            e.stopPropagation();

            const navItem = arrow.closest('.nav-item');
            if (!navItem) return;

            const isOpen = navItem.classList.toggle('open');
            arrow.setAttribute('aria-expanded', isOpen);
            saveNavState();
        });
    }

    // ========== NAV STATE PERSISTENCE ==========

    function navItemPath(item) {
        const link = item.querySelector(':scope > .nav-item-header > a');
        if (!link) return null;

        // Use pathname to get a consistent relative path from the root
        // This avoids issues with absolute vs relative URLs
        return new URL(link.href, window.location.origin).pathname;
    }

    function saveNavState() {
        const navItems = sidebar ? sidebar.querySelector('.nav-items') : null;
        if (!navItems) return;

        const alwaysOpen = navItems.dataset.alwaysOpen === 'true';
        if (alwaysOpen) {
            // When all directories are always expanded, there is no user
            // collapse/expand state to persist.
            return;
        }

        let state = {};
        try {
            state = JSON.parse(sessionStorage.getItem(NAV_STATE_STORAGE_KEY) || '{}');
        } catch (e) {
            state = {};
        }

        navItems.querySelectorAll('.nav-item.directory').forEach(item => {
            const path = navItemPath(item);
            if (!path) return;

            // Don't persist the forced-open state of the active page's
            // ancestors; that state is derived from the current page and
            // should not be treated as a user's explicit choice.
            if (item.querySelector('.nav-item.active') !== null) return;

            state[path] = item.classList.contains('open');
        });

        try {
            sessionStorage.setItem(NAV_STATE_STORAGE_KEY, JSON.stringify(state));
        } catch (e) {
            // Ignore storage errors (e.g. private mode / quota)
        }
    }

    function applyNavState(container) {
        container = container || sidebar;
        if (!container) return;

        const navItems = container.classList && container.classList.contains('nav-items')
            ? container
            : container.querySelector('.nav-items');
        const alwaysOpen = navItems && navItems.dataset.alwaysOpen === 'true';

        // In always-open mode, the server renders every directory as expanded.
        // Don't touch anything so the old behavior is preserved.
        if (alwaysOpen) return;

        let state = {};
        try {
            state = JSON.parse(sessionStorage.getItem(NAV_STATE_STORAGE_KEY) || '{}');
        } catch (e) {
            state = {};
        }

        container.querySelectorAll('.nav-item.directory').forEach(item => {
            const path = navItemPath(item);
            const hasActiveDescendant = item.querySelector('.nav-item.active') !== null;

            let isOpen;
            if (hasActiveDescendant) {
                // The active page must always be visible: expand its ancestors
                // regardless of any previously persisted collapse state.
                isOpen = true;
            } else if (path && state.hasOwnProperty(path)) {
                // Honor explicit expand/collapse choices made by the user.
                isOpen = state[path];
            } else {
                // First visit to this directory: keep it collapsed.
                isOpen = false;
            }

            item.classList.toggle('open', isOpen);
            const arrow = item.querySelector('.nav-arrow');
            if (arrow) arrow.setAttribute('aria-expanded', String(isOpen));
        });
    }

    function hydrateNavState(container) {
        const target = container || sidebar;
        if (!target) return;

        const navItems = target.classList && target.classList.contains('nav-items')
            ? target
            : target.querySelector('.nav-items');
        if (!navItems) return;

        // State restoration is not a user interaction, so suppress arrow
        // transitions until the final expanded/collapsed state is in place.
        navItems.classList.remove('nav-state-ready');
        applyNavState(target);
        requestAnimationFrame(function() {
            navItems.classList.add('nav-state-ready');
        });
    }

    // ========== SIDEBAR CORE FUNCTIONS ==========
    
    function toggleSidebar() {
        if (!hamburger || !sidebar || !body || !content) return;
        
        if (sidebar.classList.contains('active')) {
            closeSidebar();
        } else {
            openSidebar();
        }
    }

    function openSidebar() {
        if (!hamburger || !sidebar || !body || !content) return;

        // Save scroll position before position:fixed resets it
        const scrollY = window.scrollY;
        body.style.top = `-${scrollY}px`;

        hamburger.classList.add('active');
        sidebar.classList.add('active');
        body.classList.add('sidebar-active');
        content.classList.add('sidebar-active');
        
        // Clear any transforms from dragging
        resetSidebarTransforms();
    }

    function closeSidebar() {
        if (!hamburger || !sidebar || !body || !content) return;

        // Restore scroll position that was lost when position:fixed was applied
        const scrollY = -parseInt(body.style.top || '0', 10);
        body.style.top = '';

        hamburger.classList.remove('active');
        sidebar.classList.remove('active');
        body.classList.remove('sidebar-active');
        content.classList.remove('sidebar-active');

        window.scrollTo(0, scrollY);
        
        // Clear any transforms from dragging
        resetSidebarTransforms();
    }

    // ========== HAMBURGER MENU ==========
    
    function initHamburgerMenu() {
        if (!hamburger) return;

        hamburger.addEventListener('click', function(e) {
            e.stopPropagation();
            toggleSidebar();
        });
    }

    // ========== CLICK OUTSIDE TO CLOSE ==========
    
    function initClickOutside() {
        document.addEventListener('click', function(e) {
            if (sidebar &&
                sidebar.classList.contains('active') &&
                !sidebar.contains(e.target) &&
                !hamburger.contains(e.target)) {
                closeSidebar();
            }
        });
    }

    // ========== SIDEBAR LINKS ==========
    
    function initSidebarLinks() {
        if (!sidebar) return;

        sidebar.querySelectorAll('a').forEach(link => {
            link.addEventListener('click', function(e) {
                // Let the nav expand/collapse toggle handle arrow clicks
                if (e.target.closest('button.nav-arrow')) return;

                // Persist all directory states before navigation
                saveNavState();

                // Close sidebar on mobile when link is clicked
                if (window.innerWidth <= 768) {
                    closeSidebar();
                }
            });
        });
    }

    // ========== AUTO-SCROLL ACTIVE ITEM ==========
    
    function scrollActiveIntoView() {
        const navItems = document.querySelector('.nav-items');
        if (!navItems) return;

        // Find all active items
        const activeItems = navItems.querySelectorAll('.nav-item.active');
        if (!activeItems.length) return;

        // Find the deepest (most nested) active item
        let deepestItem = activeItems[0];
        let maxDepth = getElementDepth(deepestItem, navItems);

        for (let i = 1; i < activeItems.length; i++) {
            const depth = getElementDepth(activeItems[i], navItems);
            if (depth > maxDepth) {
                maxDepth = depth;
                deepestItem = activeItems[i];
            }
        }

        // Calculate offset including logo height if present
        let offset = CONFIG.autoScrollOffset;
        const logo = document.querySelector('.sidebar img, .sidebar svg');
        if (logo) {
            offset += logo.offsetHeight;
        }

        // Scroll the deepest active item into view
        navItems.scrollTop = Math.max(0, deepestItem.offsetTop - offset);
    }

    function getElementDepth(element, container) {
        let depth = 0;
        let parent = element.parentElement;

        while (parent && parent !== container) {
            depth++;
            parent = parent.parentElement;
        }

        return depth;
    }

    // ========== MOBILE TOUCH GESTURES ==========
    
    function initTouchGestures() {
        // Touch start - detect edge touches and prepare for gestures
        document.addEventListener('touchstart', handleTouchStart, { passive: false });
        
        // Touch move - real-time drag following
        document.addEventListener('touchmove', handleTouchMove, { passive: false });
        
        // Touch end - complete gesture and snap/swipe logic
        document.addEventListener('touchend', handleTouchEnd, { passive: true });
        
        // Touch cancel - cleanup
        document.addEventListener('touchcancel', handleTouchCancel, { passive: true });
    }

    function handleTouchStart(e) {
        // Only handle single touch
        if (e.touches.length !== 1) return;
        
        const touch = e.touches[0];
        touchStartX = touch.clientX;
        touchStartY = touch.clientY;
        startTime = Date.now();
        isDragging = false;
        dragProgress = 0;
        
        // Check if touch started from left edge (for opening sidebar)
        const isLeftEdgeTouch = touchStartX <= CONFIG.edgeThreshold;
        const sidebarClosed = !sidebar.classList.contains('active');
        
        // Only proceed if:
        // 1. Left edge touch when sidebar is closed, OR
        // 2. Sidebar is already open (for closing)
        if (sidebarClosed && !isLeftEdgeTouch) {
            return;
        }
        
    }

    function handleTouchMove(e) {
        if (e.touches.length !== 1) return;
        
        const touch = e.touches[0];
        const currentX = touch.clientX;
        const currentY = touch.clientY;
        
        const deltaX = currentX - touchStartX;
        const deltaY = Math.abs(currentY - touchStartY);
        
        // Exit if too much vertical movement (likely a scroll)
        if (deltaY > CONFIG.verticalThreshold) {
            return;
        }
        
        // Start drag following if moved enough horizontally
        if (Math.abs(deltaX) > CONFIG.dragFollowThreshold) {
            const isValidLeftEdgeSwipe = (touchStartX <= CONFIG.edgeThreshold && deltaX > 0 && !sidebar.classList.contains('active'));
            const isValidCloseSwipe = (sidebar.classList.contains('active') && deltaX < 0);
            
            if (isValidLeftEdgeSwipe || isValidCloseSwipe) {
                isDragging = true;
                e.preventDefault(); // Prevent scrolling during sidebar interaction
                
                // Update sidebar position in real-time
                updateSidebarPosition(deltaX);
            }
        }
    }

    function handleTouchEnd(e) {
        if (e.changedTouches.length !== 1) return;
        
        touchEndX = e.changedTouches[0].clientX;
        
        if (isDragging) {
            // Check for fast swipe during drag
            const swipeDistance = touchEndX - touchStartX;
            const swipeTime = Date.now() - startTime;
            const velocity = Math.abs(swipeDistance) / swipeTime;
            const sidebarOpen = sidebar.classList.contains('active');

            // If fast swipe, override drag threshold
            if (velocity > 0.8 && swipeTime < CONFIG.maxSwipeTime) {
                if (swipeDistance > 0 && !sidebarOpen) {
                    openSidebar();
                } else if (swipeDistance < 0 && sidebarOpen) {
                    closeSidebar();
                } else {
                    handleDragEnd();
                }
            } else {
                handleDragEnd();
            }
        } else {
            // Handle simple swipe gestures
            handleSwipeGesture();
        }
        
        // Reset state
        isDragging = false;
        dragProgress = 0;
        resetSidebarTransforms();
    }

    function handleTouchCancel() {
        isDragging = false;
        dragProgress = 0;
        resetSidebarTransforms();
    }

    // ========== DRAG FOLLOWING ==========
    
    function updateSidebarPosition(deltaX) {
        if (!sidebar) return;
        
        const isOpen = sidebar.classList.contains('active');
        const sidebarWidth = sidebar.offsetWidth;
        
        let progress;
        if (!isOpen) {
            // Opening: deltaX is positive, progress from 0 to 1
            progress = Math.max(0, Math.min(1, deltaX / sidebarWidth));
        } else {
            // Closing: deltaX is negative, progress from 1 to 0
            progress = Math.max(0, Math.min(1, 1 + (deltaX / sidebarWidth)));
        }
        
        // Update global drag progress for snap logic
        dragProgress = progress;
        
        // Apply transform to show sidebar following finger
        const translateX = isOpen ? 
            (progress - 1) * sidebarWidth : 
            (progress - 1) * sidebarWidth;
        
        sidebar.style.transform = `translateX(${translateX}px)`;
        
        // Add visual feedback shadow
        if (progress > 0.3) {
            sidebar.style.boxShadow = '2px 0 8px rgba(0, 0, 0, 0.2)';
        } else {
            sidebar.style.boxShadow = '';
        }
    }

    function resetSidebarTransforms() {
        if (!sidebar) return;
        
        sidebar.style.transform = '';
        sidebar.style.boxShadow = '';
    }

    // ========== DRAG END LOGIC ==========
    
    function handleDragEnd() {
        const isOpen = sidebar.classList.contains('active');
        
        if (isOpen) {
            // Closing: dragProgress goes from 1 (open) to 0 (closed)
            // Close if dragged enough (progress < 1 - threshold)
            if (dragProgress < (1 - CONFIG.snapThreshold)) {
                closeSidebar();
            } else {
                openSidebar(); // Snap back to open
            }
        } else {
            // Opening: dragProgress goes from 0 (closed) to 1 (open)
            if (dragProgress >= CONFIG.snapThreshold) {
                openSidebar();
            } else {
                resetSidebarTransforms(); // Snap back to closed
            }
        }
    }

    // ========== SWIPE GESTURES ==========
    
    function handleSwipeGesture() {
        if (!sidebar) return;

        const swipeDistance = touchEndX - touchStartX;
        const swipeTime = Date.now() - startTime;
        const verticalDistance = Math.abs(touchCurrentY - touchStartY);
        
        // Validate swipe gesture
        if (verticalDistance > CONFIG.verticalThreshold) return;
        if (swipeTime > CONFIG.maxSwipeTime) return;

        const sidebarOpen = sidebar.classList.contains('active');

        // Right swipe from edge to open sidebar
        if (swipeDistance > CONFIG.swipeThreshold && 
            touchStartX <= CONFIG.edgeThreshold && 
            !sidebarOpen) {
            openSidebar();
        }
        // Left swipe to close sidebar
        else if (swipeDistance < -CONFIG.swipeThreshold && sidebarOpen) {
            closeSidebar();
        }
        
        // Quick swipe detection (high velocity)
        const velocity = Math.abs(swipeDistance) / swipeTime;
        if (velocity > 0.8) {
            if (swipeDistance > 30 && touchStartX <= CONFIG.edgeThreshold && !sidebarOpen) {
                openSidebar();
            } else if (swipeDistance < -30 && sidebarOpen) {
                closeSidebar();
            }
        }
    }

    // ========== SCROLLABLE CONTENT DETECTION ==========
    
    function findScrollableParent(element) {
        if (!element || element === document.body) return null;
        
        const style = window.getComputedStyle(element);
        const isScrollable = style.overflowX === 'auto' || style.overflowX === 'scroll' ||
                           style.overflowY === 'auto' || style.overflowY === 'scroll';
        
        if (isScrollable && (element.scrollWidth > element.clientWidth || 
                           element.scrollHeight > element.clientHeight)) {
            return element;
        }
        
        // Check for known scrollable elements
        if (element.tagName === 'PRE' || element.tagName === 'CODE' || 
            element.classList.contains('CodeMirror') ||
            element.classList.contains('breadcrumbs-path')) {
            return element;
        }
        
        return findScrollableParent(element.parentElement);
    }

    // ========== SIDEBAR REFRESH ==========
    
    async function refreshSidebar() {
        if (!sidebar) return;

        try {
            const response = await fetch(window.location.pathname);
            const text = await response.text();
            const parser = new DOMParser();
            const doc = parser.parseFromString(text, 'text/html');
            const newSidebar = doc.querySelector('.nav-items');

            if (newSidebar) {
                const navItems = sidebar.querySelector('.nav-items');
                if (navItems) {
                    navItems.innerHTML = newSidebar.innerHTML;
                    initSidebarLinks(); // Reinitialize click handlers
                    hydrateNavState(navItems); // Reapply state without transition jitter
                }
            }
        } catch (error) {
            console.error('Error refreshing sidebar:', error);
        }
    }

    // Apply persisted state as early as possible when this script is loaded
    // directly after the nav tree markup, so the tree is rendered in its final
    // expanded/collapsed form before the browser paints.
    (function applyNavStateEarly() {
        const earlyNavItems = document.querySelector('.nav-items');
        if (earlyNavItems) {
            hydrateNavState(earlyNavItems);
        }
    })();

    // ========== PUBLIC API ==========

    window.SidebarNavigation = {
        toggleSidebar: toggleSidebar,
        openSidebar: openSidebar,
        closeSidebar: closeSidebar,
        refreshSidebar: refreshSidebar
    };

})();
