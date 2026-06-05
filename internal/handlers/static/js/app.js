// Snip - Client-side enhancements
(function() {
    'use strict';

    // ---- Toast Notifications ----
    function showToast(msg, duration) {
        duration = duration || 2500;
        var existing = document.querySelector('.toast');
        if (existing) existing.remove();

        var toast = document.createElement('div');
        toast.className = 'toast';
        toast.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg> ' + msg;
        document.body.appendChild(toast);
        requestAnimationFrame(function() { toast.classList.add('show'); });
        setTimeout(function() {
            toast.classList.remove('show');
            setTimeout(function() { toast.remove(); }, 300);
        }, duration);
    }
    window.showToast = showToast;

    // ---- Mobile Nav Toggle ----
    var navToggle = document.querySelector('.nav-toggle');
    var navLinks = document.querySelector('.nav-links');
    if (navToggle && navLinks) {
        navToggle.addEventListener('click', function() {
            navLinks.classList.toggle('open');
        });
        // Close on outside click
        document.addEventListener('click', function(e) {
            if (!navToggle.contains(e.target) && !navLinks.contains(e.target)) {
                navLinks.classList.remove('open');
            }
        });
    }

    // ---- Keyboard Shortcuts ----
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + Enter to submit paste form
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            var textarea = document.querySelector('.paste-textarea');
            if (textarea && document.activeElement === textarea) {
                var form = textarea.closest('form');
                if (form) {
                    e.preventDefault();
                    form.requestSubmit ? form.requestSubmit() : form.submit();
                }
            }
        }

        // / to focus search (when not in input)
        if (e.key === '/' && !isInputFocused()) {
            var search = document.querySelector('.input-search');
            if (search) {
                e.preventDefault();
                search.focus();
            }
        }

        // Escape to blur
        if (e.key === 'Escape') {
            document.activeElement.blur();
        }
    });

    function isInputFocused() {
        var el = document.activeElement;
        return el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable);
    }

    // ---- Tab Key in Textarea ----
    var pasteTextarea = document.querySelector('.paste-textarea');
    if (pasteTextarea) {
        pasteTextarea.addEventListener('keydown', function(e) {
            if (e.key === 'Tab') {
                e.preventDefault();
                var start = this.selectionStart;
                var end = this.selectionEnd;
                this.value = this.value.substring(0, start) + '    ' + this.value.substring(end);
                this.selectionStart = this.selectionEnd = start + 4;
            }
        });
    }

    // ---- QR Code (lightweight inline) ----
    // Simple QR code generator - no external dependencies
    window.genQR = function(text, size) {
        size = size || 128;
        var canvas = document.createElement('canvas');
        canvas.width = size;
        canvas.height = size;
        canvas.style.borderRadius = '8px';
        var ctx = canvas.getContext('2d');

        // Simple approach: use QR code API
        var img = new Image();
        img.crossOrigin = 'anonymous';
        img.src = 'https://api.qrserver.com/v1/create-qr-code/?size=' + size + 'x' + size + '&data=' + encodeURIComponent(text) + '&bgcolor=0d1117&color=58a6ff&format=png';
        img.onload = function() {
            ctx.drawImage(img, 0, 0, size, size);
        };
        img.style.borderRadius = '8px';
        img.width = size;
        img.height = size;
        return img;
    };

    // ---- Copy Enhancement ----
    var copyBtns = document.querySelectorAll('[data-copy]');
    copyBtns.forEach(function(btn) {
        btn.addEventListener('click', function() {
            var text = this.getAttribute('data-copy');
            navigator.clipboard.writeText(text).then(function() {
                showToast('Copied!');
            });
        });
    });

    // ---- HTMX Events ----
    document.addEventListener('htmx:afterRequest', function(e) {
        if (e.detail.successful) {
            var target = e.detail.target;
            if (target && target.id === 'result') {
                // Paste created successfully
            }
        }
    });

})();
