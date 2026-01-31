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
    formatText(text) {
        return text.replace(/\n/g, '<br>');
    }

    Connect() {
        this.socket = new WebSocket(this.url);

        this.socket.onopen = () => {
            console.log("Bağlantı Kuruldu");
            this.E_statusText.innerText = "SYSTEM ONLINE";
            this.E_statusDiv.classList.add('online');
            this.E_statusDiv.classList.remove('offline');
            this.addLog("WEB", "Uplink established with CHIMERA Core.", this.LogType.TOOL);
        };

        this.socket.onclose = () => {
            console.log("Bağlantı Koptu");
            this.E_statusText.innerText = "OFFLINE";
            this.E_statusDiv.classList.remove('online');
            this.E_statusDiv.classList.add('offline');
            this.addLog("SYSTEM", "Connection lost. Retrying in 5s...", "error");
            setTimeout(() => this.Connect(), 5000); // 5 saniye sonra tekrar dene
        };

        this.socket.onerror = (err) => {
            console.error("Socket Hatası:", err);
        };

        this.socket.onmessage = (event) => {
            try {
                this.handleServerMessage(JSON.parse(event.data));
            } catch (e) {
                console.error("JSON Parse Hatası:", e);
            }
        };

        return true;
    }

    sendMessage() {
        const text = this.E_userInput.value.trim();

        if (!text) { this.addLog("USER INPUT", "Empty input provided.", this.LogType.ERROR); return; }

        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            this.addLog("ERROR", "Sistem çevrimdışı. Mesaj gönderilemedi.", this.LogType.ERROR);
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
        // Data Tipleri: 'response' (Bot cevabı), 'thought' (Düşünce), 'tool' (Araç kullanımı), 'log' (Sistem)

        switch (data.type) {
            case 'response':
                this.addMessage(data.content, 'bot');
                break;
            case 'thought':
                this.addLog("THOUGHT", String(data.content).slice(0, 50) + "...", this.LogType.THOUGHT);
                break;
            case 'tool':
                this.addLog(`TOOL: ${data.tool_name}`, data.content, this.LogType.TOOL);
                break;
            case 'error':
                this.addLog("ERROR", data.content, this.LogType.ERROR);
                break;
            case 'system':
                this.addLog("SYSTEM", data.content, this.LogType.SYSTEM);
                break;
            default:
                console.error("Bilinmeyen mesaj tipi:", data);
        }
    }
}


function initEventListeners() {
    console.info("Initializing event listeners...");


    const sendBtn = document.getElementById('send-btn'),
        userInput = document.getElementById('user-input'),
        clearLogsBtn = document.getElementById('clear-logs'),
        logsContainer = document.getElementById('logs-container');

    sendBtn.addEventListener('click', () => { wsClient.sendMessage(); });
    userInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') wsClient.sendMessage();
    });

    clearLogsBtn.addEventListener('click', () => {
        logsContainer.innerHTML = '';
        wsClient.addLog("SYSTEM", "Logs cleared manually.", wsClient.LogType.TOOL);
    });

    return true;
}


function initSplitter() {


    const resizer = document.getElementById('dragger');
    const leftPanel = document.querySelector('.chat-panel');
    const rightPanel = document.querySelector('.logs-panel');
    const container = document.querySelector('.app-container');

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




const wsClient = new WebSocketClient(`ws://${window.location.host}/ws`);

document.addEventListener('DOMContentLoaded', () => {
    if (initSplitter()) wsClient.addLog("WEB", "1/3 Splitter initialized.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to initialize splitter.", wsClient.LogType.ERROR);

    if (initEventListeners()) wsClient.addLog("WEB", "2/3 Event listeners initialized.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to initialize event listeners.", wsClient.LogType.ERROR);

    if (wsClient.Connect()) wsClient.addLog("WEB", "3/3 WebSocket connection initiated.", wsClient.LogType.TOOL);
    else wsClient.addLog("WEB", "Failed to initiate WebSocket connection.", wsClient.LogType.ERROR);

});

