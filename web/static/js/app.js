document.addEventListener('DOMContentLoaded', () => {
    // --- ELEMENT SEÇİMLERİ ---
    const chatHistory = document.getElementById('chat-history');
    const logsContainer = document.getElementById('logs-container');
    const userInput = document.getElementById('user-input');
    const sendBtn = document.getElementById('send-btn');
    const statusText = document.getElementById('connection-status');
    const statusDiv = document.querySelector('.status');
    const clearLogsBtn = document.getElementById('clear-logs');

    // --- WEBSOCKET BAĞLANTISI ---
    // Go sunucusu 8080'de çalışacak
    let socket;
    
    function connect() {
        socket = new WebSocket("ws://localhost:8080/ws");

        socket.onopen = () => {
            console.log("Bağlantı Kuruldu");
            statusText.innerText = "SYSTEM ONLINE";
            statusDiv.classList.add('online');
            statusDiv.classList.remove('offline');
            addLog("SYSTEM", "Uplink established with CHIMERA Core.", "system");
        };

        socket.onclose = () => {
            console.log("Bağlantı Koptu");
            statusText.innerText = "OFFLINE";
            statusDiv.classList.remove('online');
            statusDiv.classList.add('offline');
            addLog("SYSTEM", "Connection lost. Retrying in 5s...", "error");
            setTimeout(connect, 5000); // 5 saniye sonra tekrar dene
        };

        socket.onerror = (err) => {
            console.error("Socket Hatası:", err);
        };

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                handleServerMessage(data);
            } catch (e) {
                console.error("JSON Parse Hatası:", e);
            }
        };
    }

    // --- MESAJ İŞLEME MANTIĞI ---
    function handleServerMessage(data) {
        // Data Tipleri: 'response' (Bot cevabı), 'thought' (Düşünce), 'tool' (Araç kullanımı), 'log' (Sistem)
        
        switch (data.type) {
            case 'response':
                addMessage(data.content, 'bot');
                break;
            case 'thought':
                addLog("THOUGHT", data.content, "thought");
                break;
            case 'tool':
                addLog(`TOOL: ${data.tool_name}`, data.content, "tool");
                break;
            case 'error':
                addLog("ERROR", data.content, "error");
                break;
            case 'system':
                addLog("SYSTEM", data.content, "system");
                break;
            default:
                console.log("Bilinmeyen mesaj tipi:", data);
        }
    }

    // --- UI FONKSİYONLARI ---
    function sendMessage() {
        const text = userInput.value.trim();
        if (!text) return;
        
        if (!socket || socket.readyState !== WebSocket.OPEN) {
            addLog("ERROR", "Sistem çevrimdışı. Mesaj gönderilemedi.", "error");
            return;
        }

        // 1. Ekrana ekle
        addMessage(text, 'user');

        // 2. Sunucuya gönder (JSON formatında)
        socket.send(JSON.stringify({
            type: "user_message",
            content: text
        }));

        // 3. Inputu temizle
        userInput.value = '';
    }

    function addMessage(text, sender) {
        const msgDiv = document.createElement('div');
        msgDiv.classList.add('message', sender);
        
        const icon = sender === 'bot' ? '<i class="fa-solid fa-robot"></i>' : '<i class="fa-solid fa-user"></i>';
        
        msgDiv.innerHTML = `
            <div class="avatar">${icon}</div>
            <div class="text">${formatText(text)}</div>
        `;
        
        chatHistory.appendChild(msgDiv);
        scrollToBottom(chatHistory);
    }

    function addLog(label, text, type) {
        const logDiv = document.createElement('div');
        logDiv.classList.add('log-entry', type);
        
        const time = new Date().toLocaleTimeString('tr-TR', { hour12: false });
        
        logDiv.innerHTML = `
            <span class="time">[${time}]</span>
            <strong>${label}:</strong> ${text}
        `;
        
        logsContainer.appendChild(logDiv);
        scrollToBottom(logsContainer);
    }

    function scrollToBottom(element) {
        element.scrollTop = element.scrollHeight;
    }

    // Basit Markdown benzeri formatlama (Opsiyonel)
    function formatText(text) {
        // Yeni satırları <br> yap
        return text.replace(/\n/g, '<br>');
    }

    // --- EVENT LISTENERS ---
    sendBtn.addEventListener('click', sendMessage);
    userInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') sendMessage();
    });
    
    clearLogsBtn.addEventListener('click', () => {
        logsContainer.innerHTML = '';
        addLog("SYSTEM", "Logs cleared manually.", "system");
    });

    // Başlat
    connect();
});