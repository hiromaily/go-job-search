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
	Jobs    []Job
}

type Job struct {
	Title        string
	MachingLevel uint8
}

func scrapeIndeed(url, country string, start int, ret chan SearchResult, wg *sync.WaitGroup) {
	lg.Debug("[URL]", fmt.Sprintf(url, start))

	doc, err := goquery.NewDocument(fmt.Sprintf(url, start))
	if err != nil {
		lg.Errorf("[scrapeIndeed]")
		return
	}

	//debug HTML
	//if url == "https://www.indeed.nl/vacatures?q=golang&start=0"{
	//	res, _ := doc.Find("body").Html()
	//	lg.Debug("[scrapeIndeed]", res)
	//}

	titles := SearchResult{Country: country}
	jobs := []Job{}

	var waitGroup sync.WaitGroup

	if start == 0{
		//<div class="resultsTop"><div id="searchCount">Vacatures 1 tot 10 van 48</div>
		searchCount := []int{}
		doc.Find("#searchCount").Each(func(_ int, s *goquery.Selection) {
			tmp := strings.Split(s.Text(), " ")
			for _, v := range tmp {
				if i, ok := strconv.Atoi(v); ok == nil {
					searchCount = append(searchCount, i)
				}
			}
		})
		lg.Debug("[searchCount]", searchCount)

		// TODO:call left pages.
		if len(searchCount) == 3{
			for i := 10; i < searchCount[2]; i += 10 {
				waitGroup.Add(1)
				go scrapeIndeed(url, country, i, ret, &waitGroup)
			}
		}
	}

	doc.Find("h2.jobtitle a").Each(func(_ int, s *goquery.Selection) {
		//lg.Debug(s)

		if title, ok := s.Attr("title"); ok {
			level := analyzeTitle(title)
			if level != 0 {
				//lg.Debug(title)
				jobs = append(jobs, Job{Title: title, MachingLevel: level})
			}
		}
	})
	titles.Jobs = jobs
	ret <- titles

	//TODO: wain until all called
	if start == 0{
		waitGroup.Wait()
	}else{
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

//func perseHTML(htmldata *goquery.Document) []string {
//}

//goroutine
func ScrapeIndeed(confs []conf.PageIndeedConfig) (ret []SearchResult) {
	resCh := make(chan SearchResult)

	var waitGroup sync.WaitGroup
	// set all request first
	waitGroup.Add(len(confs))

	// execute all request by goroutine
	for _, conf := range confs {
		go func(u, c string) {
			scrapeIndeed(u, c, 0, resCh, nil)
			waitGroup.Done()
		}(fmt.Sprintf("%s%s", conf.Url, conf.Parameter), conf.Country)
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
