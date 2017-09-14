package main

import (
	"flag"
	"fmt"
	conf "github.com/hiromaily/go-job-search/libs/config"
	enum "github.com/hiromaily/go-job-search/libs/enum"
	sc "github.com/hiromaily/go-job-search/libs/scrape"
	lg "github.com/hiromaily/golibs/log"
	tm "github.com/hiromaily/golibs/time"
	"time"
)

var (
	tomlPath = flag.String("f", "", "Toml file path")
	keyword  = flag.String("key", "", "Keyword to search")
)

func init() {
	flag.Parse()

	//log
	lg.InitializeLog(lg.DebugStatus, lg.LogOff, 99,
		"[JOB]", "/var/log/go/go-job-search.log")

	//load TOML
	conf.New(*tomlPath, false)

	//lg.Debugf("[c.Matching.Level] : %v\n", c.Matching.Level)
	//lg.Debugf("[c.Matching.Keywords] : %v\n", c.Keywords)
	//lg.Debugf("[c.Matching.Page.Indeed] : %v\n", c.Page.Indeed)
}

func main() {

	//Indeed
	callIndeed()
}

func callIndeed() {
	defer tm.Track(time.Now(), "callIndeed()")

	//scraping
	if *keyword != "" {
		conf.GetConf().Keywords[0].Search = *keyword
	}
	results := sc.ScrapeIndeed()

	// merge
	jobs := make(map[string][]sc.Job)
	// receive result channel here
	for _, result := range results {
		if _, ok := jobs[result.Country]; !ok {
			jobs[result.Country] = []sc.Job{}
		}
		for _, job := range result.Jobs {
			job.Link = result.BaseUrl + job.Link
			jobs[result.Country] = append(jobs[result.Country], job)
		}
	}

	// display
	for key, val := range jobs {
		fmt.Println("----------------------------------------")
		fmt.Printf("[Country] %s (%d)\n", enum.COUNTRY[key], len(val))
		for _, v := range val {
			fmt.Printf("[Job] %s (%s)\n", v.Title, v.Company)
			fmt.Printf("       %s\n", v.Link)
		}
	}
}
