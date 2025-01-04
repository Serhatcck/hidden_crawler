package hidden_crawler

import (
	"fmt"
	"net/url"
	"strings"
)

type Request struct {
	Method  string
	Host    string
	URL     string
	Scheme  string
	Body    string
	Headers map[string]string
}

type JsonLData struct {
	Url     string       `json:"url"`
	Request JsonLRequest `json:"request"`
}

type JsonLRequest struct {
	Headers map[string]string `json:"header"`
	Raw     string            `json:"raw"`
}

func newGetRequestFromUrl(u *url.URL) Request {
	return Request{
		URL:    u.String(),
		Method: "GET",
		Host:   u.Host,
		Scheme: u.Scheme,
	}
}

func (r Request) CreateJsonL() JsonLData {
	jsonL := JsonLData{
		Url: r.URL,
	}

	u, f := getURL(r.URL)
	if !f && u != nil && u.Host != "" {
		if r.Headers != nil {
			jsonL.Request.Headers = r.Headers
		} else {
			jsonL.Request.Headers = make(map[string]string)
		}
		jsonL.Request.Headers["host"] = u.Host
		jsonL.Request.Headers["method"] = r.Method
		jsonL.Request.Headers["path"] = u.Path
		jsonL.Request.Headers["scheme"] = u.Scheme
		jsonL.createRaw(r.Body)

	}

	return jsonL
}

func (j *JsonLData) createRaw(postData string) {
	var rawBuilder strings.Builder
	rawBuilder.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", j.Request.Headers["method"], j.Request.Headers["path"]))

	for key, value := range j.Request.Headers {
		if key == "scheme" || key == "method" || key == "path" {
			continue
		}
		rawBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	if postData != "" {
		rawBuilder.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(postData)))
		rawBuilder.WriteString("\r\n") // Header'ları bitir
		rawBuilder.WriteString(postData)
	} else {
		rawBuilder.WriteString("\r\n") // Header'ları bitir
	}

	j.Request.Raw = rawBuilder.String()

}
