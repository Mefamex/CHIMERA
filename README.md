# CHIMERA: Autonomous Self-Correcting Local Agent

**CHIMERA**, yerel kaynaklarla (Local LLM) çalışan, çok dilli (Polyglot) mikro servis mimarisine sahip, kendi kendini denetleyebilen (Self-Reflecting) ve sürekli iyileştirme prensibini benimseyen, ileri seviye bir **Yapay Zeka Ajanı ve Orkestrasyon Sistemi**dir.

Bu proje, standart bir "Chatbot" değildir. Kullanıcı isteklerini planlayan, araç (tool) kullanan, sonuçları eleştirel bir gözle denetleyen ve hatalıysa kendi kendini düzelten otonom bir "Dijital İşçi"dir.

---

## 🎯 Proje Vizyonu ve Felsefesi

Bu projenin temel amacı, küçük ve orta ölçekli Yerel Modellerin (örn: Qwen 14b, Llama 3) "Agentic" yeteneklerini, sıkı bir mimari disiplinle GPT-4 seviyesine yaklaştırmaktır.

1. **Yerellik ve Gizlilik (Local-First):** Tüm zeka (Ollama) ve araçlar yerel makinede çalışır. İnternet kesilse bile temel mantık çalışmaya devam eder.
2. **Öz Denetim (Self-Correction Loop):** Ajan, bir çıktıyı kullanıcıya sunmadan önce "Critic" (Eleştirmen) modülü ile değerlendirir. Hatalıysa, kullanıcı fark etmeden geri dönüp düzeltir.
3. **En İyi Araç Prensibi (Polyglot Architecture):** Tek bir dile saplanıp kalmaz. Hız ve yönetim için **Go**, Veri analizi için **Python**, Web otomasyonu için **Node.js** kullanılır.
4. **Şeffaf Zeka (Glass Box AI):** Ajanın sadece cevabını değil, "Düşünce Zincirini" (Chain of Thought) ve "Araç Çıktılarını" da canlı olarak gösteren bir arayüze sahiptir.

---

## 🏗️ Sistem Mimarisi

Sistem, **"Orkestra Şefi ve Uzmanlar"** modeline dayanır. Sistem bileşenleri birbirinden izole edilmiş mikro servisler (Microservices) olarak çalışır.

### 1. The Brain (Orkestratör) - `Golang`

Sistemin merkezi sinir sistemidir.

* **Görevi:** Kullanıcıdan gelen isteği alır, "Planner" (Planlayıcı) modülü ile adımlara böler ve uygun "Uzman" servise iş emri gönderir.
* **İletişim:** Diğer servislerle HTTP üzerinden, Web Arayüzü ile WebSocket üzerinden haberleşir.
* **Güvenlik:** Telegram Bot yönetimi ve Sistem Komutları (CLI) üzerindeki yetki kontrolü buradadır.

### 2. The Hands (Analiz ve Arama) - `Python`

Sistemin veri işleyen elleridir.

* **Görevi:** Arama motorlarını (Search API) kullanmak, büyük metinleri işlemek, veri analizi yapmak ve karmaşık hesaplamaları gerçekleştirmektir.
* **Özelliği:** Pydantic kütüphanesi kullanılarak Model çıktılarının (JSON) veri bütünlüğü burada zorlanır (Type Enforcement).

### 3. The Eyes (Gözlem ve Etkileşim) - `Node.js`

Sistemin internete açılan gözleridir.

* **Görevi:** Modern web sitelerini taramak (Scraping), JavaScript çalıştıran sayfalarla etkileşime girmek (Puppeteer/Playwright) ve browser tabanlı otomasyonları yönetmektir.

### 4. The Face (Kullanıcı Arayüzü) - `Vanilla Web`

Sistemin dışarıya açılan yüzüdür.

* **Yapısı:** Framework bağımsız (React/Vue yok), saf HTML/CSS/JS ile yazılmış yüksek performanslı arayüz.
* **Özelliği:** "Split-Screen" (Bölünmüş Ekran) yapısı sayesinde sol tarafta sohbet akarken, sağ tarafta Ajanın terminal çıktıları, düşünce süreci ve kullandığı araçlar canlı izlenir.

---

## 🚀 Temel Özellikler ve Yetenekler

* **Döngüsel Akıl Yürütme (ReAct Loop):** *Düşün -> Planla -> Uygula -> Gözlemle -> Tekrar Düşün.*
* **Canlı Web Erişimi:** Sadece arama sonuçlarını değil, web sitelerinin içeriğini okuyup analiz edebilir.
* **Sistem Entegrasyonu:** (Linux/Arch Özel) Terminal komutlarını çalıştırabilir, sistem bilgisini sorgulayabilir (Sandbox/İzin dahilinde).
* **Dinamik Hafıza:** Uzun süreli konuşmaları özetleyerek hafızada tutma ve bağlamı kaybetmeme.
* **Hata Yakalama:** Eğer bir araç (örn: Browser) hata verirse, Ajan bunu fark eder ve alternatif bir yol dener.

---

## 📂 Klasör Yapısı Mantığı

Bu proje **Monorepo** yapısındadır ancak her servis kendi bağımsız krallığına sahiptir:

* **`cmd/` & `internal/**`: Go dilinin standart proje düzenidir. Sistemin çekirdek mantığı (Brain) burada tutulur ve dışarıdan erişime kapalıdır.
* **`services/`**: Yan servislerin (Python ve Node) bulunduğu alandır. Her klasörün kendi paket yöneticisi (`venv`, `node_modules`) vardır. Bu sayede bağımlılık çakışması (Dependency Hell) yaşanmaz.
* **`web/`**: Go sunucusu tarafından sunulan statik frontend dosyalarıdır.
* **`configs/prompts/`**: Yapay Zekanın "Kişiliği" kodun içine gömülmez. `.txt` dosyaları olarak burada tutulur. Prompt mühendisliği buradan yönetilir.

---

## 🔄 Çalışma Akışı (Workflow)

1. **Input:** Kullanıcı Web Arayüzü veya Telegram üzerinden bir görev verir.
2. **Planlama (Go):** Orkestratör, Ollama'ya başvurarak görevi alt adımlara böler (JSON formatında).
3. **Yönlendirme:**
* Eğer görev "Araştırma" ise -> **Python Servisine** istek atılır.
* Eğer görev "Site Gezme" ise -> **Node.js Servisine** istek atılır.


4. **Uygulama:** İlgili servis görevi yapar ve sonucu ham veri olarak Go'ya döner.
5. **Değerlendirme (Critic):** Go, gelen sonucu ve kullanıcının asıl isteğini tekrar Ollama'ya gönderir: *"Bu sonuç yeterli mi?"*
6. **Karar:**
* *Yetersizse:* Döngü başa döner (Retry).
* *Yeterliyse:* Sonuç kullanıcıya doğal dilde sunulur.



---

## 🛠️ Teknoloji Yığını (Tech Stack)

* **Core Logic:** Go (Golang) 1.21+
* **Web Server:** Go Fiber & Gorilla WebSocket
* **Data Processing:** Python 3.11+ (FastAPI, Pydantic)
* **Browser Automation:** Node.js (Express, Puppeteer)
* **LLM Inference:** Ollama (Qwen 2.5 14b, Llama 3)
* **Frontend:** HTML5, CSS3, Vanilla JS (No Framework)
* **Database:** SQLite (Hafif ve yerel loglama için)

---
