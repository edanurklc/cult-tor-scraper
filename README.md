Tor Scraper, Tor ağı (The Onion Router) üzerinden yayın yapan .onion servisleri ve açık internet sitelerine anonim şekilde erişerek verileri otomatik olarak toplayan ve işleyen yazılımlardır.

CULT, Go (Golang) dili ile geliştirilmiş, tüm ağ trafiğini Tor SOCKS5 proxy üzerinden yönlendiren, hedef web sitelerine kullanıcı IP adresini ifşa etmeden bağlanabilen OSINT tabanlı ve CTI odaklı bir Tor Scraper aracıdır.

Araç, hedef web sitesinin dinamik yapısını headless Chromium ortamında yükleyerek sayfanın güncel durumunu ekran görüntüsü (screenshot) ve HTML olarak kaydeder. CULT, görsel kanıtve HTML verisi toplamaya odaklanarak low-interaction reconnaissance yaklaşımıyla çalışır.
Bu yönüyle CULT, Dark Web servislerinin ön keşfi, görsel durum tespiti ve zaman damgalı kanıt üretimi amacıyla kullanılmaktadır.

KAYNAK KODU KOPYALAYIN VE DİZİNE GİDİN
```bash
git clone https://github.com/edanurklc/cult.git
cd cult
```

TOOL'U ÇALIŞTIRIN
```bash
go run cult.go https://example.onion
```
