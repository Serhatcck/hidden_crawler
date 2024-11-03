package main

import (
	"fmt"

	"github.com/Serhatcck/hidden_crawler/pkg/hidden_crawler"
	"github.com/projectdiscovery/goflags"
)

func main() {

	var config hidden_crawler.Config

	flagSet := goflags.NewFlagSet()

	flagSet.CreateGroup("target", "Target",
		flagSet.StringVarP(&config.Url, "url", "u", "", "Target URL"),
	)

	flagSet.CreateGroup("analyze", "analyze",
		flagSet.BoolVarP(&config.CheckRobotsfile, "robots", "r", false, "Check robots.txt file"),
	)

	flagSet.CreateGroup("config", "Configuration",
		flagSet.BoolVarP(&config.FollowRedirectAnotherHost, "follow-redirect-another-host", "frah", false, "If the application redirects to a different domain, it determines whether the analysis will continue."),
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

	flagSet.SetDescription(`Web Crawler is a crawling web site for pentesting`)

	flagSet.Parse()

	hidden_crawler.BuildConf(&config)

	worker := hidden_crawler.InitWorker(&config)
	worker.Start()

	for _, req := range worker.FoundRequest {
		fmt.Println(req.URL)
	}
}
