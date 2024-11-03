package hidden_crawler

import (
	"fmt"
	"net/http"
	"strings"
)

// Robots.txt dosyasını parse eden fonksiyon
func robotstxtparser(url string, httpClient httpClient, config Config) []Request {
	// Robots.txt URL'ini oluştur
	robotsURL := strings.TrimRight(url, "/") + "/robots.txt"

	// Robots.txt dosyasına istek gönder
	req := Request{
		Method:  "GET",
		URL:     robotsURL,
		Headers: config.Headers,
	}
	resp, err := httpClient.Execute(&req)

	if err != nil {
		return nil
	}

	// Eğer robots.txt bulunamazsa boş bir liste döner
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	// Sonuçları string dizisi olarak tutacak bir slice
	var reqs []Request

	// Robots.txt içeriğini satır satır oku ve parse et
	fmt.Println()

	for _, line := range strings.Split(resp.Body, "\n") {

		// Satırda yorum varsa, yorum kısmını çıkar (# ile başlayan kısmı sil)
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		// Boş satırları atla
		if line == "" {
			continue
		}

		// Disallow veya Allow satırlarını parse et ve sadece URL'leri al
		if strings.HasPrefix(line, "Disallow:") || strings.HasPrefix(line, "Allow:") {
			// "Disallow:" veya "Allow:" kısmını çıkar ve yalnızca yolu al
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])

				// For this :
				// Disallow: */secur/forgotpassword.jsp?*
				path = strings.TrimPrefix(path, "*")

				fullURL := strings.TrimRight(url, "/") + path
				u, err := getURL(fullURL)
				if !err {
					reqs = append(reqs, newGetRequestFromUrl(u))

				}
			}
		}
	}

	return reqs
}
