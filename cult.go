package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
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

func htmlAlVeKaydet(hedefURL string) error {

	_ = os.Mkdir("html", 0755)

	ayarlar := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(torProxyAdresi),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), ayarlar...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, zamanAsimi)
	defer cancel()

	var htmlKod string

	hata := chromedp.Run(ctx,
		chromedp.Navigate(hedefURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlKod, chromedp.ByQuery),
	)
	if hata != nil {
		return hata
	}

	timestamp := time.Now().UnixNano()
	dosyaAdi := fmt.Sprintf("html/%d.html", timestamp)
	return os.WriteFile(dosyaAdi, []byte(htmlKod), 0644)
}

func torIPDogrula(logger *log.Logger) error { // bu fonksiyon yazılırken ai yardımı alınmıştır

	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9150", nil, proxy.Direct)
	if err != nil {
		return fmt.Errorf("SOCKS5 oluşturulamadı: %v", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Get("https://check.torproject.org")
	if err != nil {
		return fmt.Errorf("Tor IP doğrulama başarısız: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if strings.Contains(bodyStr, "Congratulations") {
		logger.Println("TOR IP DOĞRULAMA: BAŞARILI (Tor ağı kullanılıyor)")
		fmt.Println("TOR IP doğrulama başarılı")
		return nil
	}

	return fmt.Errorf("Tor IP doğrulama başarısız: Tor kullanılmıyor olabilir")
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
	_ = os.Mkdir("html", 0755)

	logger, logDosyasi, err := logDosyasiOlustur()
	if err != nil {
		log.Fatalf("Log dosyası oluşturulamadı: %v", err)
	}
	defer logDosyasi.Close()

	err = torIPDogrula(logger)
	if err != nil {
		logger.Println("Tor IP doğrulama başarısız:", err)
		log.Fatal(err)
	}

	hedefler, err := hedefleriOku("targets.yaml")
	if err != nil {
		log.Fatalf("targets.yaml okunamadı: %v", err)
	}

	for _, hedef := range hedefler {

		err := ekranGoruntusuAl(hedef)
		durum := hataDurumuBelirle(err)
		if err != nil {
			fmt.Printf("TARANIYOR: %s -> %s\n", hedef, durum)
			logger.Printf("BAŞARISIZ %s -> %s", hedef, durum)
			continue
		}
		fmt.Printf("TARANIYOR %s -> SUCCESS\n", hedef)
		logger.Printf("BAŞARILI %s -> SUCCESS", hedef)

		errHTML := htmlAlVeKaydet(hedef)
		durumHTML := hataDurumuBelirle(errHTML)
		if errHTML != nil {
			fmt.Printf("HTML ALINAMADI: %s -> %s\n", hedef, durumHTML)
			logger.Printf("HTML BAŞARISIZ %s -> %s", hedef, durumHTML)
			continue
		}
		logger.Printf("HTML BAŞARILI %s -> SUCCESS", hedef)
	}

	fmt.Println("Tarama işlemi tamamlandı.")
}
