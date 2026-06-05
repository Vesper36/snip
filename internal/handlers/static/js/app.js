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

    // ---- QR Code (client-side, no external dependency) ----
    // Minimal QR code generator - supports up to version 10 (alphanumeric/byte)
    window.genQR = function(text, size) {
        size = size || 128;
        var canvas = document.createElement('canvas');
        canvas.width = size;
        canvas.height = size;
        canvas.style.borderRadius = '8px';
        var qr = generateQR(text);
        if (!qr) {
            // Fallback: create a simple link text
            var span = document.createElement('span');
            span.textContent = text;
            span.style.fontSize = '0.7rem';
            span.style.wordBreak = 'break-all';
            return span;
        }
        var count = qr.length;
        var cellSize = size / count;
        var ctx = canvas.getContext('2d');
        var isLight = document.documentElement.classList.contains('light');
        ctx.fillStyle = isLight ? '#ffffff' : '#0a0e14';
        ctx.fillRect(0, 0, size, size);
        ctx.fillStyle = isLight ? '#1e293b' : '#3b82f6';
        for (var row = 0; row < count; row++) {
            for (var col = 0; col < count; col++) {
                if (qr[row][col]) {
                    ctx.fillRect(Math.floor(col * cellSize), Math.floor(row * cellSize), Math.ceil(cellSize), Math.ceil(cellSize));
                }
            }
        }
        return canvas;
    };

    // QR code generation using a compact embedded implementation
    // Based on nayuki/QR-Code-generator (simplified)
    function generateQR(text) {
        var segs = [{mode: 'byte', data: text}];
        var version = 1;
        // Find minimum version that fits the data
        for (; version <= 10; version++) {
            var dataCap = qrDataCapacity(version, 'byte');
            if (dataCap >= text.length) break;
        }
        if (version > 10) return null;
        var ecl = 'M';
        var size = version * 4 + 17;
        var modules = [], isFunction = [];
        for (var i = 0; i < size; i++) {
            modules.push(new Array(size).fill(false));
            isFunction.push(new Array(size).fill(false));
        }
        // Draw function patterns
        drawFunctionPatterns(modules, isFunction, version, size);
        // Encode data
        var dataCodewords = [];
        var byteData = [];
        for (var ci = 0; ci < text.length; ci++) byteData.push(text.charCodeAt(ci) & 0xFF);
        // Mode indicator (byte=0100)
        var charCountBits = version <= 1 ? 8 : 16;
        var bits = [];
        appendBits(bits, 4, 4); // byte mode
        appendBits(bits, byteData.length, charCountBits);
        for (var ci = 0; ci < byteData.length; ci++) appendBits(bits, byteData[ci], 8);
        // Terminator
        var totalDataBits = qrTotalDataBits(version, ecl);
        var terminatorLen = Math.min(4, totalDataBits - bits.length);
        appendBits(bits, 0, terminatorLen);
        // Pad to byte boundary
        while (bits.length % 8 !== 0) bits.push(0);
        // Pad bytes
        var totalCodewords = Math.ceil(totalDataBits / 8);
        var padBytes = [0xEC, 0x11];
        var pi = 0;
        while (Math.ceil(bits.length / 8) < totalCodewords) {
            appendBits(bits, padBytes[pi % 2], 8);
            pi++;
        }
        // Convert bits to bytes
        for (var bi = 0; bi < bits.length; bi += 8) {
            var byte = 0;
            for (var bj = 0; bj < 8; bj++) byte = (byte << 1) | (bits[bi + bj] || 0);
            dataCodewords.push(byte);
        }
        // Reed-Solomon error correction
        var result = qrGenerateCodewords(dataCodewords, version, ecl);
        if (!result) return null;
        // Draw codewords
        drawCodewords(modules, isFunction, result, size);
        // Apply mask (try mask 0 for simplicity)
        applyMask(modules, isFunction, 0, size);
        return modules;
    }

    function appendBits(arr, val, len) {
        for (var i = len - 1; i >= 0; i--) arr.push((val >>> i) & 1);
    }

    function drawFunctionPatterns(modules, isFunction, version, size) {
        // Finders
        for (var cy = 0; cy < 7; cy++) {
            for (var cx = 0; cx < 7; cx++) {
                var val = cy === 0 || cy === 6 || cx === 0 || cx === 6 || (cy >= 2 && cy <= 4 && cx >= 2 && cx <= 4);
                setFunctionModule(modules, isFunction, cx, cy, val, size);
                setFunctionModule(modules, isFunction, size - 7 + cx, cy, val, size);
                setFunctionModule(modules, isFunction, cx, size - 7 + cy, val, size);
            }
        }
        // Timing
        for (var i = 8; i < size - 8; i++) {
            setFunctionModule(modules, isFunction, 6, i, i % 2 === 0, size);
            setFunctionModule(modules, isFunction, i, 6, i % 2 === 0, size);
        }
        // Alignment patterns
        var alignPos = qrAlignmentPositions(version);
        for (var ai = 0; ai < alignPos.length; ai++) {
            for (var aj = 0; aj < alignPos.length; aj++) {
                if ((ai === 0 && aj === 0) || (ai === 0 && aj === alignPos.length - 1) || (ai === alignPos.length - 1 && aj === 0)) continue;
                for (var dy = -1; dy <= 1; dy++) {
                    for (var dx = -1; dx <= 1; dx++) {
                        setFunctionModule(modules, isFunction, alignPos[ai] + dx, alignPos[aj] + dy, dy === 0 && dx === 0 || Math.abs(dy) === 1 && Math.abs(dx) === 1 ? true : (dy === 0 && dx === 0), size);
                    }
                }
            }
        }
        // Format & version areas
        for (var fi = 0; fi < 8; fi++) {
            setFunctionModule(modules, isFunction, 8, fi, false, size);
            setFunctionModule(modules, isFunction, size - 1 - fi, 8, false, size);
            setFunctionModule(modules, isFunction, fi, 8, false, size);
            setFunctionModule(modules, isFunction, 8, size - 1 - fi, false, size);
        }
        setFunctionModule(modules, isFunction, 8, size - 8, true, size);
        if (version >= 7) {
            for (var vi = 0; vi < 6; vi++) {
                for (var vj = 0; vj < 3; vj++) {
                    setFunctionModule(modules, isFunction, size - 11 + vj, vi, false, size);
                    setFunctionModule(modules, isFunction, vi, size - 11 + vj, false, size);
                }
            }
        }
    }

    function setFunctionModule(modules, isFunction, x, y, val, size) {
        if (x >= 0 && x < size && y >= 0 && y < size) {
            modules[y][x] = val;
            isFunction[y][x] = true;
        }
    }

    function drawCodewords(modules, isFunction, data, size) {
        var i = 0;
        for (var right = size - 1; right >= 1; right -= 2) {
            if (right === 6) right = 5;
            for (var vert = 0; vert < size; vert++) {
                for (var j = 0; j < 2; j++) {
                    var x = right - j;
                    var upward = ((right + 1) & 2) === 0;
                    var y = upward ? size - 1 - vert : vert;
                    if (!isFunction[y][x] && i < data.length * 8) {
                        modules[y][x] = ((data[i >>> 3] >>> (7 - (i & 7))) & 1) !== 0;
                        i++;
                    }
                }
            }
        }
    }

    function applyMask(modules, isFunction, mask, size) {
        for (var y = 0; y < size; y++) {
            for (var x = 0; x < size; x++) {
                if (isFunction[y][x]) continue;
                var invert;
                switch (mask) {
                    case 0: invert = (x + y) % 2 === 0; break;
                    case 1: invert = y % 2 === 0; break;
                    case 2: invert = x % 3 === 0; break;
                    case 3: invert = (x + y) % 3 === 0; break;
                    case 4: invert = (Math.floor(x / 3) + Math.floor(y / 2)) % 2 === 0; break;
                    case 5: invert = (x * y) % 2 + (x * y) % 3 === 0; break;
                    case 6: invert = ((x * y) % 2 + (x * y) % 3) % 2 === 0; break;
                    case 7: invert = ((x + y) % 2 + (x * y) % 3) % 2 === 0; break;
                }
                if (invert) modules[y][x] = !modules[y][x];
            }
        }
        // Write format info
        var formatBits = qrFormatBits(mask, 'M');
        for (var fi = 0; fi < 15; fi++) {
            var bit = (formatBits >>> fi) & 1;
            if (fi < 6) modules[8][fi] = bit ? true : false;
            else if (fi < 8) modules[8][fi + 1] = bit ? true : false;
            else modules[8][size - 15 + fi] = bit ? true : false;
            if (fi < 8) modules[size - 1 - fi][8] = bit ? true : false;
            else if (fi < 9) modules[15 - fi][8] = bit ? true : false;
            else modules[15 - fi][8] = bit ? true : false;
        }
    }

    function qrFormatBits(mask, ecl) {
        var eclBits = {L: 1, M: 0, Q: 3, H: 2}[ecl];
        var data = eclBits << 3 | mask;
        var rem = data;
        for (var i = 0; i < 10; i++) rem = (rem << 1) ^ ((rem >>> 9) * 0x537);
        var bits = (data << 10 | rem) ^ 0x5412;
        return bits;
    }

    function qrDataCapacity(version, mode) {
        var ecCodewordsPerBlock = {1:10,2:16,3:26,4:18,5:24,6:16,7:18,8:22,9:22,10:26};
        var numBlocks = {1:1,2:1,3:1,4:2,5:2,6:4,7:4,8:4,9:5,10:5};
        var totalCodewords = version * 4 + 16;
        var ecPerBlock = ecCodewordsPerBlock[version];
        var totalDataCodewords = totalCodewords - ecPerBlock * numBlocks[version];
        var charCountBits = version <= 1 ? 8 : 16;
        return (totalDataCodewords * 8 - 4 - charCountBits) / 8;
    }

    function qrTotalDataBits(version, ecl) {
        var totalCodewords = version * 4 + 16;
        var numEcBlocks = {1:1,2:1,3:1,4:2,5:2,6:4,7:4,8:4,9:5,10:5};
        var ecPerBlock = {1:10,2:16,3:26,4:18,5:24,6:16,7:18,8:22,9:22,10:26};
        return (totalCodewords - ecPerBlock[version] * numEcBlocks[version]) * 8;
    }

    function qrAlignmentPositions(version) {
        if (version === 1) return [];
        var positions = [6];
        var last = version * 4 + 10;
        var step = version === 1 ? 0 : Math.ceil((last - 6) / Math.floor(version / 7 + 2));
        if (step % 2 !== 0) step++;
        var pos = last;
        while (pos > 6) { positions.unshift(pos); pos -= step; }
        positions.unshift(6);
        return positions;
    }

    function qrGenerateCodewords(dataCodewords, version, ecl) {
        var ecPerBlock = {1:10,2:16,3:26,4:18,5:24,6:16,7:18,8:22,9:22,10:26};
        var numBlocks = {1:1,2:1,3:1,4:2,5:2,6:4,7:4,8:4,9:5,10:5};
        var totalCodewords = version * 4 + 16;
        var shortDataLen = Math.floor((totalCodewords - ecPerBlock[version] * numBlocks[version]) / numBlocks[version]);
        var numLongBlocks = (totalCodewords - ecPerBlock[version] * numBlocks[version]) % numBlocks[version];
        var numShortBlocks = numBlocks[version] - numLongBlocks;
        // Split data into blocks
        var blocks = [];
        var offset = 0;
        for (var bi = 0; bi < numShortBlocks; bi++) {
            blocks.push(dataCodewords.slice(offset, offset + shortDataLen));
            offset += shortDataLen;
        }
        for (var bi = 0; bi < numLongBlocks; bi++) {
            blocks.push(dataCodewords.slice(offset, offset + shortDataLen + 1));
            offset += shortDataLen + 1;
        }
        // Generate EC for each block
        var result = [];
        // Data interleaving
        for (var di = 0; di <= shortDataLen; di++) {
            for (var bi = 0; bi < blocks.length; bi++) {
                if (di < blocks[bi].length) result.push(blocks[bi][di]);
            }
        }
        // EC interleaving
        var ecBlocks = [];
        for (var bi = 0; bi < blocks.length; bi++) {
            ecBlocks.push(reedSolomonEncode(blocks[bi], ecPerBlock[version]));
        }
        for (var di = 0; di < ecPerBlock[version]; di++) {
            for (var bi = 0; bi < ecBlocks.length; bi++) {
                result.push(ecBlocks[bi][di]);
            }
        }
        return result;
    }

    function reedSolomon(data, degree) {
        // Generator polynomial coefficients
        var gen = [1];
        for (var i = 0; i < degree; i++) {
            var root = 1;
            for (var j = 0; j < i; j++) root = gfMul(root, 2);
            var newGen = new Array(gen.length + 1).fill(0);
            for (var j = 0; j < gen.length; j++) {
                newGen[j] ^= gen[j];
                newGen[j + 1] ^= gfMul(gen[j], root);
            }
            gen = newGen;
        }
        return gen;
    }

    function reedSolomonEncode(data, degree) {
        var gen = reedSolomon([], degree);
        // Polynomial division
        var result = new Array(degree).fill(0);
        for (var i = 0; i < data.length; i++) {
            var factor = data[i] ^ result[0];
            result.shift();
            result.push(0);
            if (factor !== 0) {
                for (var j = 0; j < gen.length; j++) {
                    result[j] ^= gfMul(gen[j], factor);
                }
            }
        }
        return result;
    }

    function gfMul(a, b) {
        if (a === 0 || b === 0) return 0;
        var result = 0;
        for (var i = 7; i >= 0; i--) {
            result = (result << 1) ^ ((result >>> 7) * 0x11D);
            if ((b >>> i) & 1) result ^= a;
        }
        return result;
    }

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