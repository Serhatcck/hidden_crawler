package hidden_crawler

import (
	"fmt"
	"os"
)

const (
	Clear_Terminal = "\r\033[K"
)

func WriteStatus(jobCount int, doneJobCount int) {

	fmt.Fprintf(os.Stderr, "%s [%d/%d]", Clear_Terminal, doneJobCount, jobCount)

}
