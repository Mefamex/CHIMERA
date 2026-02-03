// file path: web/static/js/app.js

async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.socket = null;
        this.E_userInput = document.getElementById('user-input');
        this.E_statusText = document.getElementById('connection-status');
        this.E_statusDiv = document.querySelector('.status');
        this.chatHistory = document.getElementById('chat-history');
        this.logsContainer = document.getElementById('logs-container');

        this.activeBotMessages = new Map(); // msgId -> { el, text }

        this.LogType = {
            THOUGHT: 'thought',
            TOOL: 'tool',
            ERROR: 'error',
            SYSTEM: 'system',
            DEFAULT: 'default',
        };
    }

    scrollToBottom(element) {
        if (!element) return;
        try { element.scrollTop = element.scrollHeight; } catch (e) { this.addLog("WEB-ERROR", "Scroll Hatası: " + e.message, this.LogType.ERROR); }
    }
    // Escape HTML to avoid breaking DOM when rendering untrusted content
    escapeHtml(str) {
        return String(str).replace(/[&<>"'`]/g, function (s) {
            return ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;', '`': '&#96;' })[s];
        });
    }

    // Basic Markdown renderer: handles fenced code blocks, inline code, links, bold, italics
    renderMarkdown(text) {
        if (text === null || text === undefined) return '';
        // escape first
        let s = this.escapeHtml(String(text));

        // Fenced code blocks ```lang\n...```
        s = s.replace(/```(\w+)?\n([\s\S]*?)```/g, (m, lang, code) => {
            const escapedCode = this.escapeHtml(code);
            const langClass = lang ? ' language-' + lang : '';
            return `<pre><code class="${langClass}">${escapedCode}</code></pre>`;
        });

        // Inline code `code`
        s = s.replace(/`([^`]+)`/g, (m, c) => `<code class="inline-code">${c}</code>`);

        // Links [text](url)
        s = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (m, t, url) => {
            return `<a href="${url}" target="_blank" rel="noopener noreferrer">${t}</a>`;
        });

        // Bold **text** or __text__
        s = s.replace(/(\*\*|__)(.*?)\1/g, (m, p1, inner) => `<strong>${inner}</strong>`);

        // Italic *text* or _text_
        s = s.replace(/(\*|_)(.*?)\1/g, (m, p1, inner) => `<em>${inner}</em>`);

        // Preserve simple line breaks
        s = s.replace(/\n/g, '<br>');
        return s;
    }

    formatText = (text) => {
        return this.renderMarkdown(text);
    };

    Connect() {
        this.socket = new WebSocket(this.url);

        this.socket.onopen = () => {
            this.addLog("WEB", "Connection established.", this.LogType.SYSTEM);
            this.E_statusText.innerText = "SYSTEM ONLINE";
            this.E_statusDiv.classList.add('online');
            this.E_statusDiv.classList.remove('offline');
            this.addLog("WEB", "Uplink established with CHIMERA Core.", this.LogType.TOOL);
        };

        this.socket.onclose = () => {
            this.addLog("WEB", "Connection lost. Retrying in 5s...", this.LogType.ERROR);
            this.E_statusText.innerText = "OFFLINE";
            this.E_statusDiv.classList.remove('online');
            this.E_statusDiv.classList.add('offline');
            setTimeout(() => this.Connect(), 5000); // 5 saniye sonra tekrar dene
        };

        this.socket.onerror = (err) => { this.addLog("WEB-ERROR", "Socket Hatası: " + JSON.stringify(err), this.LogType.ERROR); };

        this.socket.onmessage = (event) => {
            try {
                const raw = JSON.parse(event.data);
                const payload = raw && raw.payload ? raw.payload : raw;
                this.handleServerMessage(payload, raw);
            } catch (e) { this.addLog("WEB-ERROR", "JSON Parse Hatası: " + e.message, this.LogType.ERROR); }
        };
        return true;
    }

    sendMessage() {
        const text = this.E_userInput.value.trim();

        if (!text) { this.addLog("USER INPUT", "Empty input provided.", this.LogType.ERROR); return; }

        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            this.addLog("WEB-ERROR", "Sistem çevrimdışı. Mesaj gönderilemedi.", this.LogType.ERROR);
            return;
        }
        this.addMessage(text, 'user');

        // 2. Sunucuya gönder (JSON formatında)
        this.socket.send(JSON.stringify({
            type: "user_message",
            content: text
        }));
        this.E_userInput.value = '';
        try { if (this.E_userInput.style) this.E_userInput.style.height = ''; } catch (e) { }
    }

    addMessage(text, sender) {
        const msgDiv = document.createElement('div');
        msgDiv.classList.add('message', sender);

        const icon = sender === 'bot' ? '🤖' : '👤';
        const prefix = sender === 'bot' ? 'CHIMERA-AI :' : 'USER :';

        msgDiv.innerHTML = `  <div class="avatar">${icon}</div> <div class="text"> <span class="copy-prefix">${prefix}</span> <span class="copy-content">${this.formatText(text)}</span> </div> `;

        if (!this.chatHistory) this.chatHistory = document.getElementById('chat-history');
        if (this.chatHistory) this.chatHistory.appendChild(msgDiv);
        this.scrollToBottom(this.chatHistory);
    }

    addBotMessagePlaceholder(msgId) {
        const msgDiv = document.createElement('div');
        msgDiv.classList.add('message', 'bot');
        msgDiv.dataset.msgId = msgId || '';
        msgDiv.innerHTML = `
            <div class="avatar">🤖</div>
            <div class="text">
                <span class="copy-prefix">CHIMERA-AI :</span>
                <span class="copy-content"></span>
            </div>
        `;
        if (!this.chatHistory) this.chatHistory = document.getElementById('chat-history');
        if (this.chatHistory) this.chatHistory.appendChild(msgDiv);
        this.scrollToBottom(this.chatHistory);
        if (msgId) this.activeBotMessages.set(msgId, { el: msgDiv, text: '' });
        return msgDiv;
    }

    appendBotDelta(msgId, delta) {
        if (!delta) return;
        let entry = msgId ? this.activeBotMessages.get(msgId) : null;
        if (!entry) {
            const el = this.addBotMessagePlaceholder(msgId);
            entry = { el, text: '' };
            if (msgId) this.activeBotMessages.set(msgId, entry);
        }
        entry.text += String(delta);
        const contentEl = entry.el.querySelector('.copy-content');
        if (contentEl) contentEl.innerHTML = this.formatText(entry.text);
        if (!this.chatHistory) this.chatHistory = document.getElementById('chat-history');
        this.scrollToBottom(this.chatHistory);
    }

    addLog(label, text, type) {
        // type : this.logType : 'thought', 'tool', 'error', 'system'

        const logDiv = document.createElement('div');
        const safeType = Object.values(this.LogType).includes(type) ? type : 'default';
        logDiv.classList.add('log-entry', safeType);

        const time = new Date().toLocaleTimeString('tr-TR', { hour12: false });

        const safeLabel = this.escapeHtml(label);
        const safeText = this.formatText(text);
        logDiv.innerHTML = `<span class="time">[${time}]</span> <strong>${safeLabel}:</strong> ${safeText}`;

        if (!this.logsContainer) this.logsContainer = document.getElementById('logs-container');
        if (this.logsContainer) {
            this.logsContainer.appendChild(logDiv);
            this.scrollToBottom(this.logsContainer);
        } else {
            // If logs container isn't available, silently fallback to console
            try { console.log(`${label}: ${text}`); } catch (e) { }
        }
    }

    handleServerMessage(data, envelope) {
        // data: AgentResponse benzeri {type, content, tool_name}
        // envelope: opsiyonel sarmalayıcı {target_id, broadcast, timestamp, payload}
        // Data Tipleri: 'response' (Bot cevabı), 'thought' (Düşünce), 'tool' (Araç kullanımı), 'log' (Sistem)
        if (!data || !data.type) {
            this.addLog("WEB-WARNING", "Bilinmeyen ya da boş mesaj alındı: " + JSON.stringify(data), this.LogType.TOOL);
            return;
        }
        const msgType = String(data.type).trim();
        switch (msgType) {
            case 'response':
                this.addMessage(data.content, 'bot');
                break;
            case 'response_start': {
                const msgId = envelope && envelope.id ? envelope.id : '';
                this.addBotMessagePlaceholder(msgId);
                break;
            }
            case 'response_delta': {
                const msgId = envelope && envelope.id ? envelope.id : '';
                this.appendBotDelta(msgId, data.content);
                break;
            }
            case 'response_end': {
                break;
            }
            case this.LogType.THOUGHT:
                this.addLog(this.LogType.THOUGHT.toUpperCase(), String(data.content), this.LogType.THOUGHT);
                break;
            case this.LogType.TOOL:
                this.addLog(`${this.LogType.TOOL.toUpperCase()}: ${data.tool_name}`, data.content, this.LogType.TOOL);
                break;
            case this.LogType.ERROR:
                this.addLog(this.LogType.ERROR.toUpperCase(), data.content, this.LogType.ERROR);
                break;
            case this.LogType.SYSTEM:
                this.addLog(this.LogType.SYSTEM.toUpperCase(), data.content, this.LogType.SYSTEM);
                break;
            default:
                this.addLog("UNKNOWN", `Bilinmeyen mesaj türü: ${JSON.stringify(data.type)}`, this.LogType.ERROR);
        }
    }
}

const wsClient = new WebSocketClient(`ws://${window.location.host}/ws`);

// Stacks bilgisini al ve logs'a yaz
async function fetchAndLogStacksInfo() {
    try {
        const res = await fetch('/api/stacks'),
            data = await res.json();
        async function logNested(obj, indent = "") {
            for (const [key, val] of Object.entries(obj)) {
                if (val && typeof val === 'object') {
                    wsClient.addLog("STACKS", `${indent}${key}:`, wsClient.LogType.TOOL); await sleep(50);
                    await logNested(val, indent + "\u00A0");
                } else { wsClient.addLog("STACKS", `${indent}${key}: ${val}`, wsClient.LogType.TOOL); await sleep(100); }
            }
        };
        await logNested(data);
    } catch (err) { wsClient.addLog("STACKS", "Failed to fetch stacks info.", wsClient.LogType.ERROR); }
    return true;
}


async function initEventListeners() {
    const sendBtn = document.getElementById('send-btn'),
        userInput = document.getElementById('user-input'),
        clearLogsBtn = document.getElementById('clear-logs'),
        logsContainer = document.getElementById('logs-container');

    // auto-resize helper for the textarea
    function autoResize(el) {
        if (!el) return;
        try {
            el.style.height = 'auto';
            const cs = getComputedStyle(el);
            const maxH = cs.maxHeight ? parseFloat(cs.maxHeight) : Infinity;
            const newH = Math.min(el.scrollHeight, maxH || Infinity);
            el.style.height = (newH) + 'px';
        } catch (e) { }
    }

    sendBtn.addEventListener('click', () => { wsClient.sendMessage(); autoResize(userInput); });
    // allow Enter to send (no Shift) and prevent newline; Shift+Enter inserts newline
    userInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            wsClient.sendMessage();
            autoResize(userInput);
        }
    });
    userInput.addEventListener('input', () => autoResize(userInput));
    // initialize size
    setTimeout(() => autoResize(userInput), 0);

    clearLogsBtn.addEventListener('click', () => {
        logsContainer.innerHTML = '';
        wsClient.addLog("WEB", "Logs cleared manually.", wsClient.LogType.TOOL);
    });
    return true;
}


async function initSplitter() {
    const resizer = document.getElementById('dragger'),
        leftPanel = document.querySelector('.chat-panel'),
        rightPanel = document.querySelector('.logs-panel'),
        container = document.querySelector('.row.g-0') || document.querySelector('.container-fluid');

    if (resizer && leftPanel && rightPanel && container) {
        let dragging = false;
        let rafPending = false;
        let lastEvent = null;

        const onPointerMove = (e) => {
            lastEvent = e;
            if (!dragging) return;
            if (rafPending) return;
            rafPending = true;
            requestAnimationFrame(() => {
                rafPending = false;
                if (!dragging || !lastEvent) return;
                const ev = lastEvent;
                const rect = container.getBoundingClientRect();
                const resizerW = resizer.getBoundingClientRect().width || 8;
                const x = ev.clientX;
                const min = 200;
                const max = rect.width - resizerW - 200;
                let newLeft = Math.max(min, Math.min(max, x - rect.left));
                const rightWidth = Math.max(0, Math.round(rect.width - resizerW - newLeft));
                leftPanel.style.width = newLeft + 'px';
                rightPanel.style.width = rightWidth + 'px';
            });
        };

        resizer.addEventListener('pointerdown', (e) => {
            dragging = true;
            // freeze flex so width changes take effect and prevent wrapping
            try {
                leftPanel.style.flex = '0 0 auto';
                rightPanel.style.flex = '0 0 auto';
                container.style.flexWrap = 'nowrap';
                // set current pixel width as starting point
                leftPanel.style.width = `${leftPanel.getBoundingClientRect().width}px`;
                rightPanel.style.width = `${rightPanel.getBoundingClientRect().width}px`;
            } catch (err) { }
            resizer.setPointerCapture(e.pointerId);
            document.body.classList.add('dragging');
        });
        document.addEventListener('pointermove', onPointerMove);
        document.addEventListener('pointerup', (e) => {
            dragging = false;
            try { resizer.releasePointerCapture(e.pointerId); } catch (err) { }
            try {
                leftPanel.style.flex = '';
                rightPanel.style.flex = '';
                container.style.flexWrap = '';
            } catch (err) { }
            document.body.classList.remove('dragging');
        });
        resizer.addEventListener('dblclick', () => {
            leftPanel.style.width = '55%';
            rightPanel.style.width = '45%';
        });
    }
    return true;
}

async function overrideConsole() {
    const methods = ['log', 'info', 'warn', 'error', 'debug', 'trace', 'table', 'dir', 'group', 'groupCollapsed', 'groupEnd'];
    let groupIndent = 0;
    const getCircularReplacer = () => {
        const seen = new WeakSet();
        return (key, value) => {
            if (typeof value === 'object' && value !== null) {
                if (seen.has(value)) return '[Circular]';
                seen.add(value);
            }
            return value;
        };
    };
    const safeStringify = (v) => {
        try {
            if (typeof v === 'string') return v;
            if (v === undefined) return 'undefined';
            if (typeof v === 'function') return v.toString();
            return JSON.stringify(v, getCircularReplacer());
        } catch (e) { return String(v); }
    };
    const originals = {};
    methods.forEach(m => {
        originals[m] = console[m] || function () { };
    });
    console.group = (...args) => {
        originals.group.apply(console, args);
        const msg = args.map(safeStringify).join(' ');
        try { wsClient?.addLog?.('WEB', 'GROUP: ' + msg, wsClient.LogType.DEFAULT); } catch (e) { }
        groupIndent++;
    };
    console.groupCollapsed = (...args) => {
        originals.groupCollapsed.apply(console, args);
        const msg = args.map(safeStringify).join(' ');
        try { wsClient?.addLog?.('WEB', 'GROUP(collapsed): ' + msg, wsClient.LogType.DEFAULT); } catch (e) { }
        groupIndent++;
    };
    console.groupEnd = (...args) => {
        originals.groupEnd.apply(console, args);
        try { wsClient?.addLog?.('WEB', 'GROUP END', wsClient.LogType.DEFAULT); } catch (e) { }
        if (groupIndent > 0) groupIndent--;
    };
    console.table = (tab) => {
        originals.table.call(console, tab);
        let msg;
        try {
            if (Array.isArray(tab)) msg = tab.map(r => safeStringify(r)).join('\n');
            else msg = safeStringify(tab);
        } catch (e) { msg = String(tab); }
        try { wsClient?.addLog?.('WEB', msg, wsClient.LogType.DEFAULT); } catch (e) { }
    };
    console.dir = (...args) => {
        originals.dir.apply(console, args);
        const msg = args.map(safeStringify).join(' ');
        try { wsClient?.addLog?.('WEB', msg, wsClient.LogType.DEFAULT); } catch (e) { }
    };
    ['log', 'info', 'warn', 'error', 'debug', 'trace'].forEach(m => {
        console[m] = (...args) => {
            originals[m].apply(console, args);
            const prefix = ' '.repeat(Math.max(0, groupIndent * 2));
            const msg = args.map(safeStringify).join(' ');
            try { wsClient?.addLog?.('WEB', prefix + msg, wsClient.LogType.DEFAULT); } catch (e) { }
        };
    });
    return true;
}


document.addEventListener('DOMContentLoaded', async () => {
    await sleep(100);

    if (await overrideConsole()) wsClient.addLog("WEB", "Console overridden for logging.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to override console.", wsClient.LogType.ERROR);
    await sleep(100);

    if (await fetchAndLogStacksInfo()) wsClient.addLog("WEB", "Stacks info fetched.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to fetch stacks info.", wsClient.LogType.ERROR);
    await sleep(200);

    await (() => {
        if (initSplitter()) wsClient.addLog("WEB", "1/3 Splitter initialized.", wsClient.LogType.TOOL);
        else wsClient.addLog("WEB", "Failed to initialize splitter.", wsClient.LogType.ERROR);

        if (wsClient.Connect()) wsClient.addLog("WEB", "3/3 WebSocket connection initiated.", wsClient.LogType.TOOL);
        else wsClient.addLog("WEB", "Failed to initiate WebSocket connection.", wsClient.LogType.ERROR);

        if (initEventListeners()) wsClient.addLog("WEB", "2/3 Event listeners initialized.", wsClient.LogType.TOOL);
        else wsClient.addLog("WEB", "Failed to initialize event listeners.", wsClient.LogType.ERROR);

    })();
    await sleep(100);

    wsClient.addLog("WEB", "UPLINK HAZIR. CHIMERA-AI ile etkileşime geçebilirsiniz.", wsClient.LogType.TOOL);
});

