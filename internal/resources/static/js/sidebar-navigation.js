// Sidebar Navigation Module for Wiki-Go
// Handles sidebar functionality, hamburger menu toggle, and mobile touch gestures
(function() {
    'use strict';

    // ========== MODULE STATE ==========
    let hamburger, sidebar, content, body;
    
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
        initTouchGestures();
        scrollActiveIntoView();
    });

    // ========== SIDEBAR CORE FUNCTIONS ==========
    
    function toggleSidebar() {
        if (!hamburger || !sidebar || !body || !content) return;
        
        hamburger.classList.toggle('active');
        sidebar.classList.toggle('active');
        body.classList.toggle('sidebar-active');
        content.classList.toggle('sidebar-active');
    }

    function openSidebar() {
        if (!hamburger || !sidebar || !body || !content) return;
        
        hamburger.classList.add('active');
        sidebar.classList.add('active');
        body.classList.add('sidebar-active');
        content.classList.add('sidebar-active');
        
        // Clear any transforms from dragging
        resetSidebarTransforms();
    }

    function closeSidebar() {
        if (!hamburger || !sidebar || !body || !content) return;
        
        hamburger.classList.remove('active');
        sidebar.classList.remove('active');
        body.classList.remove('sidebar-active');
        content.classList.remove('sidebar-active');
        
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
            link.addEventListener('click', function() {
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
        
        // Check if touch is on scrollable content
        const scrollableParent = findScrollableParent(e.target);
        if (scrollableParent && !isLeftEdgeTouch) {
            return; // Don't interfere with scrollable content
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
            // Handle drag-to-open/close with snap threshold
            handleDragEnd();
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
        if (dragProgress >= CONFIG.snapThreshold) {
            // Snap to open if dragged beyond threshold
            openSidebar();
        } else {
            // Snap back to closed if not dragged far enough
            resetSidebarTransforms();
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
                }
            }
        } catch (error) {
            console.error('Error refreshing sidebar:', error);
        }
    }

    // ========== PUBLIC API ==========
    
    window.SidebarNavigation = {
        toggleSidebar: toggleSidebar,
        openSidebar: openSidebar,
        closeSidebar: closeSidebar,
        refreshSidebar: refreshSidebar
    };

})();
