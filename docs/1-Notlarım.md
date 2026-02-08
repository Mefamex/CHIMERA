# NOTLAR

Bu dosya kendime hatırlatmak veya uyarıda bulunmak için bazı notlar içerir.

- [NOTLAR](#notlar)
  - [NOTLARIM](#notlarim)
  - [GÜNLÜK](#günlük)


## NOTLARIM

- Init() fonksiyonları çok sıkıntılı. şu an kullandığım yerler:
  - init config -> tüm configi ayağa kaldırıyor.
  - init logger -> Sadece global logger'ı oluşturuyor. ayar yapmıyor.

- Çalışan llm büyük mü küçük mü? bu çok önemli. çünkü tüm chain yapısı buna göre değişiyor. büyükse tek bir chain yeterli olabilir ama küçükse planlama, uygulama, değerlendirme gibi ayrı chain'ler kurmak gerekiyor. bunu nasıl çözücem bilmiyorum ama şimdilik küçükler için tasarlıyacam.

- python ve nodejs runtime'ları ayrı olarak çalışsın. node neyse de ileride yapay zeka kendisi için bir script yazmak istediğinde python runtime'ına erişmesi gerekiyor. bu da güvenlik açısından riskli olabilir. o yüzden runtime'ları birbirinden izole tutmak istiyorum. belki docker veya sandbox gibi bir şey kullanabilirim.





## Yapılacak

- main.go'dan başlayarak her fonksiyonu iyileştir.
  - Yeni log sistemine geçiş. (şimdilik trace_id ve span_id eklemeyi geç.)
  - geçtik ama tekrar kontrol edilecek. bazı kusurlarımız gün yüzüne çıktı.

- brain ve brain-engine chain yapısını iyileştir.

- Viper diye module config'i okumak için bunu kullan
- third_party çok saçma durdu. onu ileride değiştirecez.