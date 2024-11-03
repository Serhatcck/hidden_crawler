package hidden_crawler

import "net/url"

type Request struct {
	Method  string
	Host    string
	URL     string
	Schema  string
	Body    string
	Raw     string
	Headers map[string]string
}

func newGetRequestFromUrl(u *url.URL) Request {

	return Request{
		URL:    u.String(),
		Method: "GET",
		Host:   u.Host,
		Schema: u.Scheme,
	}
}
