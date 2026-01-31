CHIMERA/
├── .gitignore
├── file-structure.md
├── go.mod                     # Go modül dosyası
├── go.sum
├── Makefile                   # Tüm sistemi yöneten komut merkezi
├── README.md
│
├── 🧠 cmd/                    # [BEYİN GİRİŞİ]
│   └── agent-server/
│       └── main.go            # Orkestratör Başlatıcı
│
├── ⚙️ configs/                # [AYARLAR]
│   ├── config.yaml            # Runtime yollarının tanımlandığı yer
│   └── prompts/               # AI Kişilik dosyaları (.txt)
│
├── 💾 data/                   # [HAFIZA]
│   ├── db/                    # SQLite
│   └── logs/                  # Loglar
│
├── 🧩 internal/               # [BEYİN LOB-LARI]
│   ├── agent/                 # Karar mekanizması (Planner, Evaluator)
│   ├── clients/               # Ollama ve Tool bağlantıları
│   ├── config/                # YAML okuyucu
│   ├── models/                # Veri yapıları
│   └── server/                # HTTP/WebSocket sunucusu
│
├── ⚙️ runtimes/               # [MOTOR DAİRESİ] (Sadece kütüphaneler)
│   ├── node-v25.4.0/          # Node Motoru
│   │   ├── node_modules/      # <--- Ağır modüller BURADA yaşayacak
│   │   ├── package.json       # Tüm node araçlarının ortak listesi
│   │   └── pnpm-lock.yaml
│   │
│   └── python-3.14.2/         # Python Motoru
│       ├── venv/              # <--- Sanal ortam BURADA yaşayacak
│       └── requirements.txt   # Tüm python araçlarının ortak listesi
│
├── 🛠️ tools/                  # [ATÖLYE] (Sadece Kodlar - Dependency YOK)
│   ├── python-analyzer/       # (Eski services/python-analyzer)
│   │   └── main.py            # venv yok! Sadece saf kod.
│   │
│   └── node-browser/          # (Eski services/node-browser)
│       └── bot.js             # node_modules yok! Sadece saf kod.
│
│
├── scripts/
│   └── build.sh
│
└── 🌐 web/                    # [YÜZ]
    └── static/
        ├── index.html
        ├── css/
        └── js/





CHIMERA/
├── .gitignore
├── file-structure.md
├── go.mod
├── go.sum
├── Makefile
├── README.md
│
├── cmd/
│   └── agent-server/
│       └── main.go
│
├── configs/
│   ├── config.yaml
│   └── prompts/
│
├── data/
│   ├── db/
│   └── logs/
│
├── internal/
│   ├── agent/
│   ├── clients/
│   ├── config/
│   ├── models/
│   └── server/
│
├── runtimes/
│   ├── node-v25.4.0/
│   │   ├── node_modules/
│   │   ├── package.json
│   │   └── pnpm-lock.yaml
│   │
│   └── python-3.14.2/
│       ├── venv/
│       └── requirements.txt
│
├── tools/
│   ├── python-analyzer/
│   │   └── main.py
│   │
│   ├── node-browser/
│   │   └── bot.js
│   │
│   └── scripts/
│
└── web/
    └── static/
        ├── index.html
        ├── css/
        └── js/