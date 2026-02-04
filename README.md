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

## 🚀 Temel Özellikler ve Yetenekler

* **Döngüsel Akıl Yürütme (ReAct Loop):** *Düşün -> Planla -> Uygula -> Gözlemle -> Tekrar Düşün.*
* **Canlı Web Erişimi:** Sadece arama sonuçlarını değil, web sitelerinin içeriğini okuyup analiz edebilir.
* **Sistem Entegrasyonu:** (Linux/Arch Özel) Terminal komutlarını çalıştırabilir, sistem bilgisini sorgulayabilir (Sandbox/İzin dahilinde).
* **Dinamik Hafıza:** Uzun süreli konuşmaları özetleyerek hafızada tutma ve bağlamı kaybetmeme.
* **Hata Yakalama:** Eğer bir araç (örn: Browser) hata verirse, Ajan bunu fark eder ve alternatif bir yol dener.

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

## 🤖 Asistan-Odaklı Rehber

Bu kısım, projenin "asistan" (agent) perspektifinden tam bir özetidir: yetenekler, mimari bileşenler, hafıza modeli, güvenlik ve geliştirme öncelikleri. Amaç: gelecekte bu repoyu okuyan bir geliştiricinin veya kullanıcıyın agent'in ne yaptığını, nerede olduğunu ve nasıl genişletileceğini hızlıca anlamasıdır.

**Asistan Önceliği**
- CHIMERA bir "asistan platformu"dur; ana hedef agent'in yetenekleri, yetki sınırları, hafızası ve güvenlik politikalarıdır. Kurulum talimatlarından ziyade agent davranışı ve genişletilebilirlik odaklıdır.

**Bir Bakışta Yetenekler (Skills)**
- Sistem kontrolü: `skills` modülleri aracılığıyla dosya/klasör yönetimi, süreç başlatma/durdurma, paket kurulumları.
- Terminal komutları: `cmd` modülü ile shell/terminal komutları çalıştırma ve çıktıyı yakalama.
- Klavye kontrolü: `keyboard` yeteneği ile otomatik metin yazma, tuş kombinasyonları ve kısa yollar.
- Dosya işlemleri: oluşturma, okuma, düzenleme, kopyalama, silme.
- Yazılım yönetimi: paket yükleme, servis başlatma/durdurma, derleme işlemleri.
- Kod yazma/çeşitlendirme: yeni proje/dosya oluşturma ve temel kod düzenleme (scaffold).
- İnternet araştırması: tarama, scraping ve veri çekme (Node.js puppeteer/playwright ile).
- Mesaj uygulamaları: Telegram/WhatsApp/Signal için modüler entegrasyon (isteğe bağlı, ayrı modüller).
- Kişiselleştirilmiş öğrenme: kullanıcı davranışını izleyip tercihleri hafızada tutma.
- Hatırlatıcılar & takvim: yerel hatırlatıcılar ve zamanlanmış görevler.
- Kendini geliştirme: yeni `skills` ekleyebilecek plug-in altyapısı.

**Mimari (Agent Perspektifi) — Kısa**
- Orkestratör (Brain) — `Go`: `cmd/` ve `internal/` içindeki `agent.Brain` agent'in merkezi; LLM çağrıları, planlama ve karar döngüsü burada yürütülür. (bkz: [internal/agent/brain.go](internal/agent/brain.go))
- LLM İstemcisi — `internal/clients/ollama`: Ollama ile sohbet/streaming iletişimi sağlar. (bkz: [internal/clients/ollama/client.go](internal/clients/ollama/client.go))
- Web Arayüzü — `web/static`: canlı gözlem, onay ve etkileşim için split-screen arayüz.
- WS & API — `cmd/agent-server/main.go` HTTP ve WebSocket giriş noktaları ve uygulama başlatma.
- Python servisleri: arama, veri işleme ve analiz için ayrılmış mikroservisler.
- Node.js servisleri: tarayıcı otomasyonu ve JS-uygulama etkileşimleri (puppeteer/playwright).

**Hafıza & Veri**
- Kalıcı hafıza dizini: `data/db/` — SQLite veya benzeri hafif DB planlanmıştır.
- Örnek tablo seti: `users`, `memories`, `reminders`, `skills_meta`.
- Gizlilik: Kişisel veriler şifrelenmeli; erişim sadece yerel ve yetkili süreçlere verilmeli.
- Kısa dönem kontekst: WebSocket/bağlantı başına RAM tabanlı konuşma geçmişi; uzun vadeli bağlam gerektiğinde özetleme yapıp DB'ye kaydetme.

**İletişim Kanalları**
- Yerel Web UI: ana kontrol ve manuel onay noktası.
- Mesaj uygulamaları: Telegram/WhatsApp/Signal — her biri adapter şeklinde izole edilerek eklenmeli.

**Prompt Mühendisliği & LLM**
- Prompt'lar `configs/prompts/` dizininde tutulur; agent persona ve başlangıç talimatları buradan yönetilir.
- `Brain` streaming ve normal chat metodlarıyla LLM'den cevap alır (bkz: [internal/agent/brain.go](internal/agent/brain.go)).

**Güvenlik & İzinler (Kritik)**
- Agent yalnızca sahibinin makinesinde çalışacak şekilde sınırlandırılmalıdır.
- Tehlikeli yetenekler (`allow_cmd`, `allow_keyboard`, `allow_network`) policy ile açıkça kontrol edilmeli.
- API anahtarları ve hassas konfigürasyonlar dosyaya düz metin yazılmamalı; çevre değişkenleri veya şifrelenmiş saklama tercih edilmeli.

**Öncelikli Yapılacaklar (Roadmap)**
1. İzin & güvenlik katmanının (policy engine) uygulanması. (yüksek öncelik)
2. `skills` altyapısının modüler hale getirilmesi — plugin/adapter mimarisi.
3. Kalıcı hafıza: SQLite şema tasarımı, migration ve CRUD API'leri.
4. Telegram entegrasyonu: temel POC (güvenli token yönetimi ile).
5. Browser automation POC (Node.js + puppeteer) ve hata tolerans mekanizmaları.

**Geliştirme Kuralları / Prensipler**
- Her yeni yetenek küçük, izole bir modül olarak eklenmeli; `Brain` sadece yönlendirme ve güvenlik kararlarını vermeli.
- Kullanıcı izni gerektiren eylemler için manuel onay veya policy tabanlı otomasyon zorunlu olmalı.
- Loglar `data/logs/` altında tutulmalı; gerektiğinde anonimleştirilebilmeli.

**Notlar ve İleri Adımlar**
- Bu belge asistan-odaklı rehber olarak güncellenecektir. İstersen, bu içeriği `docs/DEV.md` altında daha detaylı geliştirici rehberine dönüştürebilirim.
