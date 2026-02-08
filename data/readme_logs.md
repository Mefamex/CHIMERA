
# CHIMERA — Loglar

Bu dizin CHIMERA uygulamasının çalışma zamanı loglarını içerir

Kısa kullanım:
- Konum: data/logs/
- Format: JSONL (olaylar, hata yığınları, istek/yanıt izleri)
- Amaç: hata ayıklama, performans ve olay takibi

Bakım notu: Log hacmi büyükse periyodik temizleme veya retention politikası uygulayın. Daha fazla yapılandırma için `configs/config.yaml` ve `internal/core/logger.go` dosyalarına bakabilirsiniz.

## Log Yapısı
```
logs/
└── log_YYYY_MM_DD_HH_MM_SS/
    ├── system.jsonl
    ├── errors.jsonl
    ├── chain.jsonl
    ├── chat.jsonl
    ├── trace_brain.jsonl
    ├── trace_llmbot.jsonl
    ├── tools.jsonl
    ├── internals.jsonl
    ├── internal_python.jsonl
    └── internal_node.jsonl
```

- `system.jsonl`       : Altyapı Sağlığı, Başlatma (Startup), Port bağlama, DB bağlantısı, Kapanma (Shutdown/Signal).
- `errors.jsonl`       : Acil Müdahale, Stack trace, panik durumları. Hızlı debug için ilk bakılacak yer.
- `chain.jsonl`        : Akış İskeleti, Verisiz, sadece kimin kimle konuştuğu, Sonsuz döngüleri (Loop) tespit etmek için
- `chat.jsonl`         : Kullanıcı Vitrini, Temiz, okunabilir konuşma geçmişi, Düşünce zinciri buraya girmez.
- `trace_brain.jsonl`  : Bağlam & Hafıza. Sistem promptu + Hafıza + Kullanıcı girdisi + Plan. Ajanın neden o kararı verdiğini açıklar.
- `trace_llmbot.jsonl` : Ham LLM Verisi. JSON parse hatalarını ve LLM halüsinasyonlarını yakalamak için (Input/Output dump).
- `tools.jsonl`        : Araç Kullanımı & Süreler, Fonksiyon adı, parametreler, başarı durumu, çağrı süresi. Performans izleme ve darboğaz tespiti için.
- `internals.jsonl`    : Go'nun alt servislere verdiği emirler.
- `internal_python.jsonl`  : The Hands - stdout ve stderr.
- `internal_node.jsonl`: The Eyes - stdout ve stderr.


## Loglarda Gerekli Parametreler

- `system.jsonl`: time, level, component, msg, ...
- `errors.jsonl`: time, level, component, culprit, function , stacktrace, msg, ...
- `chain.jsonl`: time, level, component, from, action, (status), msg, ...
- `chat.jsonl`: time, level, component, sender, msg, ...
- `trace_brain.jsonl`: time, level, component, phase, msg, ...
- `trace_llmbot.jsonl`: time, level, component, model, duration, raw_prompt, msg, ...
- `tools.jsonl`: time, level, component, tool_name, args, output, success, duration, msg, ...
- `internals.jsonl`: time, level, component, target, endpoint, response, cmd, msg, ...
- `internal_python.jsonl`: time, level, component, script, response, msg ...