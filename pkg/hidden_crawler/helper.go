package hidden_crawler

import (
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func getExtension(u string) string {
	parsedURL, _ := url.Parse(u)
	segments := strings.Split(parsedURL.Path, "/")
	lastSegment := segments[len(segments)-1]
	ext := path.Ext(lastSegment)
	return ext
}

func getURL(str string) (*url.URL, bool) {
	u, err := url.Parse(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, true
	}
	return u, false
}

func getHost(str string) string {
	u, err := url.Parse(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Host
}

func compareQueries(oldUrl string, newUrl string) bool {
	oldParsed, err := url.Parse(oldUrl)
	if err != nil {
		return false
	}
	newParsed, err := url.Parse(newUrl)
	if err != nil {
		return false
	}

	// Query parametrelerini map olarak alıyoruz
	oldParams := oldParsed.Query()
	newParams := newParsed.Query()

	//Parametre yok ise
	if len(oldParams) == 0 || len(newParams) == 0 {
		return false
	}

	// Key'lerin eşleşip eşleşmediğini kontrol ediyoruz
	oldKeys := make(map[string]struct{})
	for key := range oldParams {
		oldKeys[key] = struct{}{}
	}

	newKeys := make(map[string]struct{})
	for key := range newParams {
		newKeys[key] = struct{}{}
	}

	// Key'lerin aynı olup olmadığını kontrol ediyoruz
	return reflect.DeepEqual(oldKeys, newKeys)
}

// mergeURL iki parametre alır: url ve baseUrl. Eğer url bir path ise, baseUrl ile birleştirir.
func mergeURL(urlStr string, baseUrl string) (*url.URL, error) {
	// URL'nin tam olup olmadığını kontrol et
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {

		base, err := url.Parse(baseUrl)
		if err != nil {
			return nil, err
		}

		// Path'i base URL ile birleştir
		base.Path = path.Join(base.Path, urlStr)
		return base, nil
	}

	return parsedUrl, nil

}

func getChromedpNavigateTask(target string) chromedp.Tasks {
	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "en-US,en;q=0.9",
		})), // Ekstra header'ları ayarla
		chromedp.Navigate(target),       // Hedef URL
		chromedp.WaitReady("body"),      // Wait for the body to be fully loaded
		chromedp.Sleep(1 * time.Second), // Yükleme için bekleme
	}
}
