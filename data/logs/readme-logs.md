
# CHIMERA — Loglar

Bu dizin CHIMERA uygulamasının çalışma zamanı loglarını içerir. Log dosyaları genelde JSON formatındadır ve adlarında zaman damgası bulunur (örnek: chimera-2026-02-04-03-28-37.json).

Kısa kullanım:
- Konum: data/logs/
- Format: JSON (olaylar, hata yığınları, istek/yanıt izleri)
- Amaç: hata ayıklama, performans ve olay takibi

Bakım notu: Log hacmi büyükse periyodik temizleme veya retention politikası uygulayın. Daha fazla yapılandırma için `configs/config.yaml` ve `internal/core/logger.go` dosyalarına bakabilirsiniz.

