package hidden_crawler

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var msgChann = make(chan cdproto.Message)

type headlessClient struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	config   Config
	FormUrls []string
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
		chromedp.Flag("headless", conf.Headless),
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
						return

					}
				}
			}
		}
	})

	formFiller := &FormFiller{}

	err := chromedp.Run(ctx, getChromedpNavigateTask(target))
	if err != nil {
		fmt.Println(err)
		return headlessResponse
	}

	var htmlLinks []string

	htmlLinks, err = formFiller.GetaHref(ctx)

	for _, url := range htmlLinks {
		u, err := mergeURL(url, target)
		if err == nil {
			headlessResponse.HtmlLinks = append(headlessResponse.HtmlLinks, newGetRequestFromUrl(u))
		}
	}

	formFiller.SendForms(ctx, hlClient, target)

	return headlessResponse
}

func (hlClient headlessClient) newHeadlessNetworkRequest(req *network.EventRequestWillBeSent) Request {
	var newReq = Request{
		Method: req.Request.Method,
		URL:    req.Request.URL,
	}
	if req.Request.HasPostData && len(req.Request.PostDataEntries) > 0 {
		firstEntry := req.Request.PostDataEntries[0].Bytes
		decodedBytes, err := base64.StdEncoding.DecodeString(firstEntry)
		if err != nil {
			//fmt.Println("Base64 decoding error:", err)
		}
		newReq.Body = string(decodedBytes)
	}
	if len(req.Request.Headers) > 0 {
		newReq.Headers = convertHeaders(req.Request.Headers)
	}

	return newReq
}

func convertHeaders(input map[string]interface{}) map[string]string {
	output := make(map[string]string)
	for key, value := range input {
		// Değerin string olup olmadığını kontrol et
		strValue, ok := value.(string)
		if !ok {
			return nil
		}
		output[key] = strValue
	}
	return output
}

func (hlClient *headlessClient) isFormUniqe(formUrl string) bool {
	for _, url := range hlClient.FormUrls {
		if url == formUrl {
			return false
		}
	}
	hlClient.addFormUrl(formUrl)
	return true
}

func (hlClient *headlessClient) addFormUrl(formUrl string) {
	hlClient.FormUrls = append(hlClient.FormUrls, formUrl)
}
