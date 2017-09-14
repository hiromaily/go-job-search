package scrape

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	conf "github.com/hiromaily/go-job-search/libs/config"
	lg "github.com/hiromaily/golibs/log"
	"strconv"
	"strings"
	"sync"
)

type SearchResult struct {
	Country string
	BaseUrl string
	Jobs    []Job
}

type Job struct {
	Title        string
	Link         string
	MachingLevel uint8
}

func scrapeIndeed(url, param, country, keyword string, start int, ret chan SearchResult, wg *sync.WaitGroup) {
	//lg.Debug("[URL]", fmt.Sprintf(url+"/"+param, keyword,start))

	doc, err := goquery.NewDocument(fmt.Sprintf(url+"/"+param, keyword, start))
	if err != nil {
		lg.Errorf("[scrapeIndeed]")
		return
	}

	//debug HTML
	//if url == "https://www.indeed.nl/vacatures?q=golang&start=0"{
	//	res, _ := doc.Find("body").Html()
	//	lg.Debug("[scrapeIndeed]", res)
	//}

	titles := SearchResult{Country: country, BaseUrl: url}
	jobs := []Job{}

	var waitGroup sync.WaitGroup

	//paging
	if start == 0 {
		searchCount := []int{}
		doc.Find("#searchCount").Each(func(_ int, s *goquery.Selection) {
			tmp := strings.Split(s.Text(), " ")
			for _, v := range tmp {
				if i, ok := strconv.Atoi(v); ok == nil {
					searchCount = append(searchCount, i)
				}
			}
		})
		//lg.Debug("[searchCount]", searchCount)

		// call left pages.
		if len(searchCount) == 3 {
			for i := 10; i < searchCount[2]; i += 10 {
				waitGroup.Add(1)
				go scrapeIndeed(url, param, country, keyword, i, ret, &waitGroup)
			}
		}
	}

	//analyze title
	doc.Find("h2.jobtitle a").Each(func(_ int, s *goquery.Selection) {

		link, _ := s.Attr("href")

		if title, ok := s.Attr("title"); ok {
			level := analyzeTitle(title)
			if level != 0 {
				//lg.Debug(title)
				jobs = append(jobs, Job{Title: title, Link: link, MachingLevel: level})
			}
		}
	})
	titles.Jobs = jobs
	ret <- titles

	//wait until all called
	if start == 0 {
		waitGroup.Wait()
	} else {
		wg.Done()
	}
}

func analyzeTitle(title string) uint8 {
	//lg.Debug(title)
	if strings.Index(title, "Golang") != -1 ||
		strings.Index(title, "Go ") != -1 ||
		strings.Index(title, "Go,") != -1 ||
		strings.Index(title, "Go)") != -1 {
		return 1
	}

	return 0
}

//goroutine
func ScrapeIndeed() (ret []SearchResult) {
	c := conf.GetConf()
	resCh := make(chan SearchResult)

	var waitGroup sync.WaitGroup
	// set all request first
	waitGroup.Add(len(c.Page.Indeed))

	//TODO: to deal with multiple keyworks

	// execute all request by goroutine
	for _, conf := range c.Page.Indeed {
		go func(u, p, c, key string) {
			scrapeIndeed(u, p, c, key, 0, resCh, nil)
			waitGroup.Done()
		}(conf.Url, conf.Param, conf.Country, c.Keywords[0].Search)
	}

	// close channel when finishing all goroutine
	go func() {
		waitGroup.Wait()
		close(resCh)
	}()

	// wait until results channel is closed.
	for result := range resCh {
		//lg.Infof("[ScrapeIndeed] result: %v\n", result)
		ret = append(ret, result)
	}
	return
}
