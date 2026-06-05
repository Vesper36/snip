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
        document.addEventListener('click', function(e) {
            if (!navToggle.contains(e.target) && !navLinks.contains(e.target)) {
                navLinks.classList.remove('open');
            }
        });
    }

    // ---- Keyboard Shortcuts ----
    document.addEventListener('keydown', function(e) {
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
        if (e.key === '/' && !isInputFocused()) {
            var search = document.querySelector('.input-search');
            if (search) {
                e.preventDefault();
                search.focus();
            }
        }
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

    // ---- Drag & Drop File Upload ----
    var dropZone = document.getElementById('drop-zone');
    var fileInput = document.getElementById('file-input');
    var textarea = document.querySelector('.paste-textarea');

    if (dropZone && fileInput) {
        ['dragenter', 'dragover'].forEach(function(ev) {
            dropZone.addEventListener(ev, function(e) {
                e.preventDefault();
                e.stopPropagation();
                dropZone.classList.add('drag-over');
            });
        });
        ['dragleave', 'drop'].forEach(function(ev) {
            dropZone.addEventListener(ev, function(e) {
                e.preventDefault();
                e.stopPropagation();
                dropZone.classList.remove('drag-over');
            });
        });

        // Drop on the whole form
        var form = document.querySelector('.paste-form');
        if (form) {
            form.addEventListener('dragover', function(e) {
                e.preventDefault();
                dropZone.classList.add('drag-over');
            });
            form.addEventListener('dragleave', function(e) {
                if (!form.contains(e.relatedTarget)) {
                    dropZone.classList.remove('drag-over');
                }
            });
            form.addEventListener('drop', function(e) {
                e.preventDefault();
                dropZone.classList.remove('drag-over');
                var files = e.dataTransfer.files;
                if (files.length > 0) {
                    handleFile(files[0]);
                }
            });
        }

        fileInput.addEventListener('change', function() {
            if (this.files.length > 0) {
                handleFile(this.files[0]);
            }
        });
    }

    function handleFile(file) {
        var reader = new FileReader();
        reader.onload = function(e) {
            if (textarea) {
                textarea.value = e.target.result;
                textarea.removeAttribute('required');
                // Update title with filename
                var titleInput = document.querySelector('.input-title');
                if (titleInput && !titleInput.value) {
                    titleInput.value = file.name;
                }
                // Set language to plaintext for files
                var langSelect = document.querySelector('.select-lang');
                if (langSelect) {
                    langSelect.value = 'plaintext';
                }
                showToast('File loaded: ' + file.name);
            }
        };
        reader.readAsText(file);
    }

    // ---- QR Code ----
    window.genQR = function(text, size) {
        size = size || 128;
        var img = new Image();
        img.crossOrigin = 'anonymous';
        img.src = 'https://api.qrserver.com/v1/create-qr-code/?size=' + size + 'x' + size + '&data=' + encodeURIComponent(text) + '&bgcolor=0a0e14&color=3b82f6&format=png';
        img.style.borderRadius = '8px';
        img.width = size;
        img.height = size;
        return img;
    };

    // ---- Copy Enhancement ----
    document.querySelectorAll('[data-copy]').forEach(function(btn) {
        btn.addEventListener('click', function() {
            var text = this.getAttribute('data-copy');
            navigator.clipboard.writeText(text).then(function() {
                showToast('Copied!');
            });
        });
    });

    // ---- Theme Toggle ----
    window.toggleTheme = function() {
        var html = document.documentElement;
        var isLight = html.classList.contains('light');
        if (isLight) {
            html.classList.remove('light');
            html.classList.add('dark');
            localStorage.setItem('theme', 'dark');
            var hljs = document.getElementById('hljs-css');
            if (hljs) hljs.href = 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css';
        } else {
            html.classList.remove('dark');
            html.classList.add('light');
            localStorage.setItem('theme', 'light');
            var hljs2 = document.getElementById('hljs-css');
            if (hljs2) hljs2.href = 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css';
        }
    };

    // ---- Image Paste Support ----
    if (textarea) {
        textarea.addEventListener('paste', function(e) {
            var items = e.clipboardData ? e.clipboardData.items : null;
            if (!items) return;
            for (var i = 0; i < items.length; i++) {
                if (items[i].type.indexOf('image') !== -1) {
                    e.preventDefault();
                    var blob = items[i].getAsFile();
                    var reader = new FileReader();
                    reader.onload = function(ev) {
                        textarea.value = '[Image: ' + blob.name + ']\n' + ev.target.result;
                        showToast('Image pasted as data URL');
                    };
                    reader.readAsDataURL(blob);
                    return;
                }
            }
        });
    }

    // ---- Hit Counter ----
    // Animate stat values on settings page
    document.querySelectorAll('.stat-value').forEach(function(el) {
        var target = parseInt(el.textContent, 10);
        if (isNaN(target) || target === 0) return;
        var start = 0;
        var duration = 800;
        var startTime = null;
        function animate(ts) {
            if (!startTime) startTime = ts;
            var progress = Math.min((ts - startTime) / duration, 1);
            var eased = 1 - Math.pow(1 - progress, 3);
            el.textContent = Math.round(eased * target);
            if (progress < 1) requestAnimationFrame(animate);
        }
        requestAnimationFrame(animate);
    });

})();