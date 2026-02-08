# NOTLAR

Bu dosya kendim için tuttuğum projedeki gelişmeleri, öğrendiklerimi ve hatırlatmaları içerir.
İleride kendimi geliştiriken yaptığım hataları, öğrendiğim dersleri ve önemli notları hatırlamak için buraya yazıyorum.
Proje gidişatını da net şekilde görebilmek için günlük tutuyorum.
Biraz geç oldu ama artık düzenli olarak günlüğe yazmaya çalışcam.


### 2026-02-08
- Günlük ve not tutmaya başladım.

- Şimdiye kadar öğrendiklerim:
  - Go mükemmel bişe ya... daha önce öğrenmiştim ve küçük scriptler yazmıştım ama orta ölçekli bir projede temel dil olarak hiç kullanmamıştım. modülerlik yapısı o kadar güzel ki kafa karışıklığı hiç yaratmıyor. sadece en başta modülerlik yapısında fonksiyonlar, public-private yapısına alışmam zordu.

- Şimdiye kadar yaptıklarım:
  - /cmd/agent : main.go çalışarak proje başlıyor.
  - /configs : config.yaml ve yapay zekaya verilecek prompt'lar burada duruyor.
  - /internal/core : core init: config, logger, initRuntimes
  - /internal/ollama : client'ı oluşturdum. go'nun kendi ollama'sını kullandım. ileride diğer providerlarda ortak kullanabileceğim bir interface yapısı kuracam.
  - /internal/brain : agent'in beyni. burada planlama, araç çağırma, değerlendirme vs. yapacak. ama geçici olarak burayı yapay zekaya yaptırmıştım. patlıyacağını biliyordum bugün düzelcem belki.
  - /internal/skills : burada agent'in kullanabileceği araçları tanımladım: readfile, writefile, runcmd. cmd ile pc'yi zaten kontrol eder istediği gibi de neyse... buraya ileride browser, telegram, whatsapp, instagram, searchEngine falan gelicek.
  - /internal/server: web server'ı burada kontrol ediliyor: static, ping, ws var.
  - /web/static -> yerel web ui: chatting ve logları görüntüleme
  - /runtimes : şimdilik burada python ve nodejs var.
  - /scripts : build, downloadthird_partys, generateTree var.

- Yapılacaklar:
  - Loglamaları geliştirmek:
    - şuan sadece bazı kritikleri, hataları, olayları, tool çağırmayı logluyorum.
    - yeni: sistem, brain, allchat, tooluse ayrı ayrı loglayacam.
  - Brain'i düzeltmek:
    - şuan brain'i yapay zekaya yazdırmıştım. ama çok saçma şeyler yapıyor. planlama, uygulama, değerlendirme gibi ayrı ayrı chain'ler kurmam gerekiyor. ayrıca planlama kısmında da çok saçma şeyler yapıyor. mesela "Bu görevi yapmak için hangi araçları kullanmalıyım?" diye sorduğumda "Araştırma yapmak için Python servisini kullanmalıyım, site gezmek için Node.js servisini kullanmalıyım" gibi saçma cevaplar veriyor.
  - Tool kullanmak için sadece json çıktı ver dediğimde IQ'su çok düşüyor ve bazen önce düşüncesini sonra json'u veriyor. bu yüzden şu loglamayı halledip ona göre useTool'u şekillendirecem.

- Başladık:
  - proje ilerlememi yapay zekalar ile tartıştığımda çok yetersiz cevaplar veriyordu. Geniş düşünemiyorlar. Hala yapay zekalar cok dar... yani bu projede tekim. Ve bu günlük de yalnızlığımı götürecek :D
  - main.go dan başlayıp tek tek modülleri kontrol ediyorum.
    - "CHIMERA" ascii art'ını ekledim. güzel oldu.
    - initCore -> loadConfig -> compareYaml (yeni) : yaml dosyası ve struct'ı karşılaştırarak eksik veya fazla alanları kontrol ediyor. bu sayede config dosyasındaki hataları daha kolay bulabileceğim.
      - config için yeni yapı: log-/ içinde ayrı dosyalarda sistemin logları alınacak.
  - `/scripts/stacks_info.go` move `/internal/common/stacksInfo.go` : `common.StacksInfo.Get()` return `common.StacksInfo.SystemStatus`
  - (new) `/scripts/custom_tree.go` -> folder_tree.txt : dosya ağacını asci ile yazdırır
  - (new) `cmd/logViewer/main.go` : logları görüntülemek için basit webserver
  - (new) `web/static/viewLogs.html` : logları görüntülemek için basit bir arayüz.
  - yeni log yapısını entegre ettim. çok uzun sürdü...
  - (new) `cmd/log-gui/main.go` : websocket ile logları canlı olarak görüntüleyebiliyoruz. ama performans sorunları var. Bu yüzden yeni app'de izole bir app yaptım. logları bu uygulama ile görüntülüyorum.
  - bu 2 log görüntüleme çok güzel oldu. şimdilik bunları geçtim.
  - normal chat ekranının yanındaki logs penceresine de tüm logları göstericek şekilde ayarladım.
  - brain ve brain-engine'yi halletmenin zamanı geldi...



