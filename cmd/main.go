package main

import (
	"flag"
	conf "github.com/hiromaily/go-job-search/libs/config"
	sc "github.com/hiromaily/go-job-search/libs/scrape"
	lg "github.com/hiromaily/golibs/log"
)

var (
	tomlPath = flag.String("f", "", "Toml file path")
)

func init() {
	flag.Parse()

	//log
	lg.InitializeLog(lg.DebugStatus, lg.LogOff, 99,
		"[JOB]", "/var/log/go/go-job-search.log")

}

func main() {
	//load TOML
	c := conf.New(*tomlPath, false)

	//lg.Debugf("conf %#v\n", conf.GetConfInstance())
	lg.Debugf("[c.Matching.Level] : %v\n", c.Matching.Level)
	lg.Debugf("[c.Matching.Keywords] : %v\n", c.Keywords)
	lg.Debugf("[c.Matching.Page.Indeed] : %v\n", c.Page.Indeed)

	//scraping
	titles := sc.ScrapeIndeed(c.Page.Indeed)
	lg.Infof("[titles] %v", titles)
}
