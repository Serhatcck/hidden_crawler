package hidden_crawler

import (
	"slices"
	"sync"
	"time"
)

type Worker struct {
	//WorkQueue    []WorkQueue
	Targets          chan string
	TargetList       []string
	FoundRequest     []Request
	Config           *Config
	httpClient       *httpClient
	parser           parser
	waitGroup        sync.WaitGroup
	jobCount         int
	doneJobCount     int
	jobCountChan     chan int
	doneJobCountChan chan int
	isrunning        bool
}

func (w *Worker) appendNewTarget(req Request) {

	for _, target := range w.TargetList {

		if !w.compareUrls(target, req.URL) {
			return
		}
	}
	w.jobCountChan <- 1
	w.appendFoundRequest(req)
	w.TargetList = append(w.TargetList, req.URL)
	go func() { w.Targets <- req.URL }()
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
	for _, newReq := range req {
		if w.Config.UseScope {
			if !w.isInScope(newReq.URL) {
				continue
			}
		}

		var matcher = false
		for _, oldReq := range w.FoundRequest {
			if !w.compareUrls(oldReq.URL, newReq.URL) {
				if oldReq.Method == newReq.Method {
					matcher = true
					continue
				}
			}

		}
		if !matcher {
			w.appendFoundRequest(newReq)
		}
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
	w.doneJobCountChan = make(chan int)
	w.jobCountChan = make(chan int)
	w.isrunning = true
	w.doneJobCount = 0
	w.jobCount = 0
	go w.analyzeProgress()
	w.run()
}

func (w *Worker) analyzeProgress() {
	if !w.Config.Silent {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Her bir saniyede bir toplam gorutin sayısını ekrana bas
				//goroutinesPerSecondChan <- reqPerSecondCount
				// counter'ı sıfırla
				if !w.isrunning {
					return
				}

				WriteStatus(w.jobCount, w.doneJobCount)

				if w.jobCount == w.doneJobCount {
					w.close()
				}

			case alpha := <-w.jobCountChan:
				w.jobCount += alpha

			case delta := <-w.doneJobCountChan:
				w.doneJobCount += delta

			}
		}
	}
}

func (w *Worker) run() {
	hlClient := newHeadlessClient(w.Config)
	threadLimiter := make(chan struct{}, w.Config.Threads) // Buffered channel to limit to 10 threads

	for _, url := range w.Config.Targets {
		go func() {
			w.Targets <- url
			w.jobCountChan <- 1
			if w.Config.CheckRobotsfile {
				reqs := robotstxtparser(url, *w.httpClient, *w.Config)
				for _, req := range reqs {
					w.appendNewTarget(req)
				}
			}
		}()
	}

	for target := range w.Targets {
		w.waitGroup.Add(1)
		threadLimiter <- struct{}{}

		go func(url string) {
			headlessResp := hlClient.analyseWebPage(url)
			//w.parseNetworkRequest(requests.NetworkRequest)
			newRequests := w.parser.parseRequests(headlessResp.NetworkRequest)
			w.appendFoundRequestBatch(newRequests)
			newHtmlReq := w.parser.parseRequests(headlessResp.HtmlLinks)

			for _, htmlUrl := range newHtmlReq {
				w.appendNewTarget(htmlUrl)
			}
			w.doneJobCountChan <- 1
			w.waitGroup.Done()
			<-threadLimiter

		}(target)
	}
}

func (w *Worker) close() {
	w.waitGroup.Wait()
	close(w.Targets)
}
