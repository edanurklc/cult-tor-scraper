package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	torProxyAdresi = "socks5://127.0.0.1:9150" 
	zamanAsimi     = 45 * time.Second
)


func ekranGoruntusuAl(hedefURL string) error {

	ayarlar := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(torProxyAdresi),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	ayiriciContext, iptal := chromedp.NewExecAllocator(context.Background(), ayarlar...)
	defer iptal()

	tarayiciContext, iptal := chromedp.NewContext(ayiriciContext)
	defer iptal()

	tarayiciContext, iptal = context.WithTimeout(tarayiciContext, zamanAsimi)
	defer iptal()

	var goruntu []byte

	hata := chromedp.Run(
		tarayiciContext,
		chromedp.Navigate(hedefURL),
		chromedp.Sleep(5*time.Second),
		chromedp.FullScreenshot(&goruntu, 90),
	)

	if hata != nil {
		return hata
	}

	dosyaAdi := fmt.Sprintf("screenshots/%d.png", time.Now().UnixNano())
	return os.WriteFile(dosyaAdi, goruntu, 0644)
}


func hedefleriOku(dosyaYolu string) ([]string, error) {

	dosya, hata := os.Open(dosyaYolu)
	if hata != nil {
		return nil, hata
	}
	defer dosya.Close()

	var hedefler []string
	okuyucu := bufio.NewScanner(dosya)

	for okuyucu.Scan() {
		satir := strings.TrimSpace(okuyucu.Text())
		if satir != "" {
			hedefler = append(hedefler, satir)
		}
	}

	return hedefler, nil
}


func logDosyasiOlustur() (*log.Logger, *os.File, error) {

	dosya, hata := os.Create("log_kayıtları.log")
	if hata != nil {
		return nil, nil, hata
	}

	logger := log.New(dosya, "", log.LstdFlags)
	return logger, dosya, nil
}



func hataDurumuBelirle(hata error) string {

	if hata == nil {
		return "SUCCESS"
	}

	hataMetni := hata.Error()

	switch {
	case strings.Contains(hataMetni, "deadline exceeded"):
		return "TIMEOUT"
	case strings.Contains(hataMetni, "ERR_PROXY_CONNECTION_FAILED"):
		return "PROXY_ERROR"
	default:
		return "FAILED"
	}
}



func main() {

	_ = os.Mkdir("screenshots", 0755)

	logger, logDosyasi, hata := logDosyasiOlustur()
	if hata != nil {
		log.Fatalf("Log dosyası oluşturulamadı: %v", hata)
	}
	defer logDosyasi.Close()

	hedefler, hata := hedefleriOku("targets.yaml")
	if hata != nil {
		log.Fatalf("targets.yaml okunamadı: %v", hata)
	}

	for _, hedef := range hedefler {

		hata := ekranGoruntusuAl(hedef)
		durum := hataDurumuBelirle(hata)

		if hata != nil {
			fmt.Printf("TARANIYOR: %s -> %s\n", hedef, durum)
			logger.Printf("BAŞARISIZ %s -> %s", hedef, durum)
			continue
		}

		fmt.Printf("TARANIYOR %s -> SUCCESS\n", hedef)
		logger.Printf("BAŞARILI %s -> SUCCESS", hedef)
	}

	fmt.Println("Tarama işlemi tamamlandı.")
}
