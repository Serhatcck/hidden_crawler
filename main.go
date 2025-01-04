package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Serhatcck/hidden_crawler/pkg/hidden_crawler"
	"github.com/projectdiscovery/goflags"
)

func main() {

	var config hidden_crawler.Config

	flagSet := goflags.NewFlagSet()

	flagSet.CreateGroup("target", "Target",

		flagSet.StringVarP(&config.Url, "url", "u", "", "Target URL"),
	)

	flagSet.CreateGroup("headless", "Headless",
		flagSet.BoolVarP(&config.Headless, "headless", "hd", true, "Headles Mode"),
	)

	flagSet.CreateGroup("analyze", "analyze",

		flagSet.BoolVarP(&config.CheckRobotsfile, "robots", "r", false, "Check robots.txt file"),
	)

	flagSet.CreateGroup("config", "Configuration",

		flagSet.BoolVarP(&config.FollowRedirectAnotherHost, "follow-redirect-another-host", "frah", false, "If the application redirects to a different domain, it determines whether the analysis will continue."),
		flagSet.IntVarP(&config.Threads, "thread", "t", 10, "Number of Thread"),
	)

	flagSet.CreateGroup("scope", "Scope",

		//TO DO
		flagSet.VarP(&config.ScopeTargetsStrings, "scope-target", "st", "Only give hostname"),
		flagSet.BoolVarP(&config.UseScope, "scope-use", "sc", true, "Allow scope"),
		flagSet.BoolVarP(&config.UniqueParameters, "unique-params", "up", false, ""),
	)

	flagSet.CreateGroup("filter", "filter",

		flagSet.VarP(&config.FilterExtensionsStrings, "fiter-extension", "fe", ".png,.css,.jpeg etc..."),
		flagSet.BoolVarP(&config.FilterImages, "filter-images", "fi", true, ""),
	)

	flagSet.CreateGroup("output", "output",
		flagSet.StringVarP(&config.FileName, "output", "o", "", "Output JSON File name"),
	)

	flagSet.SetDescription(`Web Crawler is a crawling web site for pentesting`)

	flagSet.Parse()

	hidden_crawler.BuildConf(&config)

	worker := hidden_crawler.InitWorker(&config)
	worker.Start()

	file, err := os.Create(config.FileName)

	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	for _, req := range worker.FoundRequest {
		jsonL := req.CreateJsonL()

		// Struct'ı JSON'a dönüştür
		jsonData, err := json.Marshal(jsonL)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			continue
		}

		// JSON verisini dosyaya yaz
		_, err = file.Write(jsonData)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			continue
		}

		// Satır sonu ekle
		_, err = file.WriteString("\n")
		if err != nil {
			fmt.Println("Error writing newline to file:", err)
			continue
		}
	}

	defer file.Close()

}
