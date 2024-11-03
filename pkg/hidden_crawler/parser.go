package hidden_crawler

import (
	"regexp"
	"slices"
	"strings"
)

type parser struct {
	conf Config
}

func NewParser(conf Config) parser {
	var parser parser
	parser.conf = conf
	return parser
}

func (p parser) parseRequests(reqs []Request) []Request {
	var newReqList []Request
	//TO DO: mailto:*
	for _, req := range reqs {
		if strings.HasPrefix(req.URL, "data:") || strings.HasPrefix(req.URL, "mailto:") {
			continue
		}

		if slices.Contains(p.conf.FilterExtensions, getExtension(req.URL)) {
			continue
		}

		re := regexp.MustCompile(`#.*?\?`)
		// '#' ve '?' arasını sildikten sonra '?' karakterini geri ekler
		req.URL = re.ReplaceAllString(req.URL, "?")

		re = regexp.MustCompile(`#.*$`)
		req.URL = re.ReplaceAllString(req.URL, "")

		newReqList = append(newReqList, req)
	}

	return newReqList
}
