package hidden_crawler

import (
	"fmt"
	"os"
)

const (
	Clear_Terminal = "\r\033[K"
)

func WriteStatus(target string, jobCount int, doneJobCount int) {

	//fmt.Fprintf(os.Stderr, "%s [%d/%d]", Clear_Terminal, doneJobCount, jobCount)
	fmt.Fprintf(os.Stderr, "%s[Target: %s] -- [Crawled Page: %d] -- [Target Page: %d]", Clear_Terminal, target, doneJobCount, jobCount)
}
