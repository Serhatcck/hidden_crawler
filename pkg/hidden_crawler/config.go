package hidden_crawler

import (
	"errors"
	"fmt"
	"strings"
)

type headerFlags map[string]string

type customParamFlags []string

type Config struct {
	Targets                   []string
	Url                       string
	FileName                  string
	CheckRobotsfile           bool
	ProxyUrl                  string
	TimeOut                   int
	Headers                   map[string]string
	CustomHeaders             headerFlags
	FollowRedirectAnotherHost bool
	UseScope                  bool
	ScopeTargets              []string
	ScopeTargetsStrings       customParamFlags
	FilterExtensions          []string
	FilterExtensionsStrings   customParamFlags
	UniqueParameters          bool
	FilterImages              bool
	Threads                   int
	Silent                    bool
	Headless                  bool
	OutputFile                string
	MaxCrawlingSource         int
}

func BuildConf(conf *Config) error {

	conf.ScopeTargets = conf.ScopeTargetsStrings

	conf.Targets = append(conf.Targets, conf.Url)
	for _, url := range conf.Targets {
		targetUrl, err := getURL(url)
		if err {
			return errors.New("given url isn't url")
		}
		conf.ScopeTargets = append(conf.ScopeTargets, targetUrl.Host)

	}

	conf.Headers = make(map[string]string)
	for key, value := range conf.CustomHeaders {
		conf.Headers[key] = value
	}

	conf.FilterExtensions = conf.FilterExtensionsStrings

	if conf.FilterImages {
		conf.FilterExtensions = append(conf.FilterExtensions, constImageExtension...)
	}

	if conf.Headers["User-Agent"] == "" {
		conf.Headers["User-Agent"] = "User-Agent Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:130.0) Gecko/20100101 Firefox/130.0"
	}

	return nil
}

func (h *headerFlags) String() string {
	return fmt.Sprint(*h)
}

func (h *headerFlags) Set(value string) error {
	if *h == nil {
		*h = make(map[string]string)
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format, expecting key:value, got %s", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	(*h)[key] = val
	return nil
}

func (f *customParamFlags) Set(value string) error {

	values := strings.Split(value, ",")
	*f = append(*f, values...)
	return nil
}

func (f *customParamFlags) String() string {
	return fmt.Sprint(*f)
}
