package hidden_crawler

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var msgChann = make(chan cdproto.Message)

type headlessClient struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	config   Config
}

type headlessResponse struct {
	NetworkRequest []Request
	HtmlLinks      []Request
}

func newHeadlessClient(conf *Config) headlessClient {
	var headlessClient headlessClient
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"

	// Başlatma için bir context oluşturun
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.IgnoreCertErrors,
		chromedp.UserAgent(userAgent),
		chromedp.Flag("disable-features", "MediaRouter"),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("deny-permission-prompts", true),
		chromedp.Flag("redirect", false),
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	headlessClient.allocCtx = allocCtx
	headlessClient.cancel = cancel
	headlessClient.config = *conf

	return headlessClient
}

func (hlClient *headlessClient) analyseWebPage(target string) headlessResponse {
	var headlessResponse headlessResponse
	ctx, cancel := chromedp.NewContext(hlClient.allocCtx)
	defer cancel()

	// Ağ izlemeyi etkinleştirin ve tüm istekleri yakalayın
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if req, ok := ev.(*network.EventRequestWillBeSent); ok {

			headlessResponse.NetworkRequest = append(headlessResponse.NetworkRequest, hlClient.newHeadlessNetworkRequest(req))
			// İstek URL'sini yakalayın
			//fmt.Println(req.Request.Headers)

			// Eğer istek bir POST ise, body'i yazdırın
			//if req.Request.Method == "POST" {
			//	fmt.Printf("POST İstek Gövdesi: %s\n", req.Request.PostDataEntries)
			//}
		}
	})

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			//Redirect check
			if ev.Type == "Document" {
				if !hlClient.config.FollowRedirectAnotherHost {
					newUrl, _ := url.Parse(ev.Response.URL)
					oldUrl, _ := url.Parse(target)
					if newUrl.Host != oldUrl.Host {
						cancel()

					}
				}
			}
		}
	})

	var htmlLinks []string

	err := chromedp.Run(ctx,
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "en-US,en;q=0.9",
			"Custom-Header":   "MyCustomValue", // İsteğe bağlı özel header
		})), // Ekstra header'ları ayarla
		chromedp.Navigate(target),     // Hedef URL
		chromedp.Sleep(5*time.Second), // Yükleme için bekleme
		chromedp.Evaluate(`
		(() => {
			const urls = [];
			document.querySelectorAll('a[href]').forEach(el => {
				if (el.href) urls.push(el.href);
				if (el.src) urls.push(el.src);
			});
			return urls;
		})()
	`, &htmlLinks),
	)

	for _, url := range htmlLinks {
		u, err := mergeURL(url, target)
		if err == nil {
			headlessResponse.HtmlLinks = append(headlessResponse.HtmlLinks, newGetRequestFromUrl(u))
		}
	}

	if err != nil {
		fmt.Println(err)
	}

	return headlessResponse
}

func (hlClient headlessClient) newHeadlessNetworkRequest(req *network.EventRequestWillBeSent) Request {
	var newReq = Request{
		Method: req.Request.Method,
		URL:    req.Request.URL,
		//Schema
		//Host
	}

	return newReq
}
