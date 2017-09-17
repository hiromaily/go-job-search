package main

import (
	"flag"
	"fmt"
	conf "github.com/hiromaily/go-job-search/libs/config"
	"github.com/hiromaily/go-job-search/libs/enum"
	sc "github.com/hiromaily/go-job-search/libs/scrape"
	lg "github.com/hiromaily/golibs/log"
	tm "github.com/hiromaily/golibs/time"
	"sort"
	"time"
)

var (
	tomlPath = flag.String("f", "", "Toml file path")
	target   = flag.Int("target", 0, "Target of search")
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
	//scraping
	if *keyword != "" {
		conf.GetConf().Keywords[0].Search = *keyword
	}

	c := conf.GetConf()

	switch *target {
	case 0:
		//Indeed
		start(c.Page.Indeed, 1)
		//Stackoverflow
		start(c.Page.Stackoverflow, 2)
		//Linkedin
		start(c.Page.Linkedin, 3)
	case 1:
		//Indeed
		start(c.Page.Indeed, 1)
	case 2:
		//Stackoverflow
		start(c.Page.Stackoverflow, 2)
	case 3:
		//Linkedin
		start(c.Page.Linkedin, 3)
	default:
	}

}

func start(pages []conf.PageConfig, mode int) {
	defer tm.Track(time.Now(), fmt.Sprintf("start():%s", enum.MODE[mode]))

	// scrape
	results := sc.Scrape(pages, mode)

	// merge
	jobs := make(map[string][]sc.Job)
	for _, result := range results {
		if _, ok := jobs[result.Country]; !ok {
			jobs[result.Country] = []sc.Job{}
		}
		for _, job := range result.Jobs {
			job.Link = result.BaseUrl + job.Link
			jobs[result.Country] = append(jobs[result.Country], job)
		}
	}
	//remove duplicated url
	for key := range jobs {
		sort.Slice(jobs[key], func(i, j int) bool {
			return jobs[key][i].Link < jobs[key][j].Link
		})

		saved := ""
		for i, v := range jobs[key] {
			if i != 0 && saved == v.Link {
				jobs[key][i].Link = ""
				continue
			}
			saved = v.Link
		}
	}

	//sort by company
	for key := range jobs {
		sort.Slice(jobs[key], func(i, j int) bool {
			return jobs[key][i].Company < jobs[key][j].Company
		})
	}

	// display
	fmt.Printf("================= %s =================\n", enum.MODE[mode])

	var country string
	for key, val := range jobs {
		fmt.Println("----------------------------------------")
		if val, ok := enum.COUNTRY[key]; ok {
			country = val
		} else {
			country = key
		}

		fmt.Printf("[Country] %s (%d)\n", country, len(val))
		for _, v := range val {
			if v.Link != "" {
				fmt.Printf("[Job] %s (%s)\n", v.Title, v.Company)
				fmt.Printf("       %s\n", v.Link)
			}
		}
	}
	fmt.Println("----------------------------------------")
}
