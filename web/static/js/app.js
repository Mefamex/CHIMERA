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

        this.LogType = {
            THOUGHT: 'thought',
            TOOL: 'tool',
            ERROR: 'error',
            SYSTEM: 'system',
            DEFAULT: 'default',
        };
    }

    scrollToBottom(element) {
        element.scrollTop = element.scrollHeight;
    }
    formatText = (text) => text.replace(/\n/g, '<br>');

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
    }

    addMessage(text, sender) {
        const msgDiv = document.createElement('div');
        msgDiv.classList.add('message', sender);

        const icon = sender === 'bot' ? '<i class="fa-solid fa-robot"></i>' : '<i class="fa-solid fa-user"></i>';

        msgDiv.innerHTML = `
            <div class="avatar">${icon}</div>
            <div class="text">${this.formatText(text)}</div>
        `;

        this.chatHistory.appendChild(msgDiv);
        this.scrollToBottom(this.chatHistory);
    }

    addLog(label, text, type) {
        // type : this.logType : 'thought', 'tool', 'error', 'system'

        const logDiv = document.createElement('div');
        const safeType = Object.values(this.LogType).includes(type) ? type : 'default';
        logDiv.classList.add('log-entry', safeType);

        const time = new Date().toLocaleTimeString('tr-TR', { hour12: false });

        logDiv.innerHTML = `
            <span class="time">[${time}]</span>
            <strong>${label}:</strong> ${text}
        `;

        this.logsContainer.appendChild(logDiv);
        this.scrollToBottom(this.logsContainer);
    }

    handleServerMessage(data) {
        // data: AgentResponse benzeri {type, content, tool_name}
        // envelope: opsiyonel sarmalayıcı {target_id, broadcast, timestamp, payload}
        // Data Tipleri: 'response' (Bot cevabı), 'thought' (Düşünce), 'tool' (Araç kullanımı), 'log' (Sistem)
        if (!data || !data.type) {
            this.addLog("WEB-WARNING", "Bilinmeyen ya da boş mesaj alındı: " + JSON.stringify(data), this.LogType.TOOL);
            return;
        }
        switch (data.type) {
            case 'response':
                this.addMessage(data.content, 'bot');
                break;
            case this.LogType.THOUGHT:
                this.addLog(this.LogType.THOUGHT.toUpperCase(), String(data.content).slice(0, 50) + "...", this.LogType.THOUGHT);
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
                this.addLog("UNKNOWN", `Bilinmeyen mesaj türü: ${data.type}`, this.LogType.ERROR);
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
                    wsClient.addLog("STACKS", `${indent}${key}:`, wsClient.LogType.TOOL); await sleep(100);
                    await logNested(val, indent + "\u00A0\u00A0");
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

    sendBtn.addEventListener('click', () => { wsClient.sendMessage(); });
    userInput.addEventListener('keypress', (e) => { if (e.key === 'Enter') wsClient.sendMessage(); });

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
        container = document.querySelector('.app-container');

    if (resizer && leftPanel && rightPanel && container) {
        let dragging = false;

        const onPointerMove = (e) => {
            if (!dragging) return;
            const rect = container.getBoundingClientRect();
            const x = e.clientX;
            const min = 200;
            const max = rect.width - 200;
            let newLeft = Math.max(min, Math.min(max, x - rect.left));
            const percent = (newLeft / rect.width) * 100;
            leftPanel.style.width = percent + '%';
            rightPanel.style.width = (100 - percent) + '%';
        };

        resizer.addEventListener('pointerdown', (e) => {
            dragging = true;
            resizer.setPointerCapture(e.pointerId);
            document.body.classList.add('dragging');
        });
        document.addEventListener('pointermove', onPointerMove);
        document.addEventListener('pointerup', (e) => {
            dragging = false;
            try { resizer.releasePointerCapture(e.pointerId); } catch (err) { }
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
    ['log', 'info', 'warn', 'error'].forEach(m => {
        const orig = console[m];
        console[m] = (...args) => {
            orig.apply(console, args);
            const msg = args.map(a => typeof a === 'object' ? JSON.stringify(a, null, 2) : a).join(' ');
            wsClient.addLog("WEB", msg, wsClient.LogType.DEFAULT);
        };
    });
    return true;
}


document.addEventListener('DOMContentLoaded', async () => {
    await sleep(100);

    if (await overrideConsole()) wsClient.addLog("WEB", "Console overridden for logging.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to override console.", wsClient.LogType.ERROR);
    await sleep(500);

    if (await fetchAndLogStacksInfo()) wsClient.addLog("WEB", "Stacks info fetched.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to fetch stacks info.", wsClient.LogType.ERROR);
    await sleep(500);

    if (await initSplitter()) wsClient.addLog("WEB", "1/3 Splitter initialized.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to initialize splitter.", wsClient.LogType.ERROR);
    await sleep(500);

    if (await initEventListeners()) wsClient.addLog("WEB", "2/3 Event listeners initialized.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to initialize event listeners.", wsClient.LogType.ERROR);
    await sleep(500);

    if (await wsClient.Connect()) wsClient.addLog("WEB", "3/3 WebSocket connection initiated.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to initiate WebSocket connection.", wsClient.LogType.ERROR);
});

