package hidden_crawler

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

type Worker struct {
	//WorkQueue    []WorkQueue
	Targets      chan string
	TargetList   []string
	FoundRequest []Request
	Config       *Config
	httpClient   *httpClient
	parser       parser
	waitGroup    sync.WaitGroup
}

func (w *Worker) appendNewTarget(req Request) {

	for _, target := range w.TargetList {

		if !w.compareUrls(target, req.URL) {
			return
		}
	}
	w.appendFoundRequest(req)
	w.TargetList = append(w.TargetList, req.URL)
	w.Targets <- req.URL
}

// if newUrl is usefull func will be return true
// if newUrl is not usefull func will be return false
func (w *Worker) compareUrls(oldUrl string, newUrl string) bool {
	if oldUrl == newUrl {
		return false
	}

	if w.Config.UniqueParameters {
		if compareQueries(oldUrl, newUrl) {
			return false
		}
	}

	if w.Config.UseScope {
		if w.isInScope(newUrl) {
			return true
		}
	}

	return false
}

func (w *Worker) isInScope(newUrl string) bool {
	u, err := getURL(newUrl)
	if err {
		return false
	}

	return slices.Contains(w.Config.ScopeTargets, u.Host)
}

func (w *Worker) appendFoundRequest(req Request) {
	w.FoundRequest = append(w.FoundRequest, req)
}

func (w *Worker) appendFoundRequestBatch(req []Request) {
	req = w.parser.parseRequests(req)
	var newReqList []Request
	for _, newReq := range req {
		if !w.isInScope(newReq.URL) {
			continue
		}
		var matcher = false
		for _, oldReq := range w.FoundRequest {

			if !w.compareUrls(oldReq.URL, newReq.URL) {
				continue
			}
			matcher = true

		}
		if !matcher {
			newReqList = append(newReqList, newReq)
		}
	}

	for _, req := range newReqList {
		w.appendFoundRequest(req)
	}
}

func InitWorker(conf *Config) *Worker {
	return &Worker{
		Config: conf,
	}
}

func (w *Worker) Start() {
	w.httpClient = newHttpClient(w.Config)
	w.parser = NewParser(*w.Config)
	w.Targets = make(chan string)
	w.run()
}

func (w *Worker) run() {
	hlClient := newHeadlessClient(w.Config)

	for _, url := range w.Config.Targets {
		go func() {
			w.Targets <- url
			if w.Config.CheckRobotsfile {
				reqs := robotstxtparser(url, *w.httpClient, *w.Config)
				for _, req := range reqs {
					w.appendNewTarget(req)
				}
			}
		}()
	}
	go func() {
		time.Sleep(5 * time.Second)
		w.waitGroup.Wait()
		close(w.Targets)
	}()

	for target := range w.Targets {
		fmt.Println("Target: ", target)
		w.waitGroup.Add(1)
		go func(url string) {
			headlessResp := hlClient.analyseWebPage(url)
			//w.parseNetworkRequest(requests.NetworkRequest)
			newRequests := w.parser.parseRequests(headlessResp.NetworkRequest)
			w.appendFoundRequestBatch(newRequests)
			newHtmlReq := w.parser.parseRequests(headlessResp.HtmlLinks)

			for _, htmlUrl := range newHtmlReq {
				w.appendNewTarget(htmlUrl)
			}
			w.waitGroup.Done()
		}(target)

	}

}
