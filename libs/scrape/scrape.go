package scrape

import (
	conf "github.com/hiromaily/go-job-search/libs/config"
	//lg "github.com/hiromaily/golibs/log"
	ur "net/url"
	"strings"
	"sync"
)

type Scraper interface {
	scrape(int, chan SearchResult, *sync.WaitGroup)
}

type SearchResult struct {
	Country string
	BaseUrl string
	Jobs    []Job
}

type Job struct {
	Title        string
	Company      string
	City         string
	Link         string
	MachingLevel uint8
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

func callScraper(s Scraper, resCh chan SearchResult, wg *sync.WaitGroup) {
	s.scrape(0, resCh, nil)
	wg.Done()
}

func encode(url string) string {
	u := &ur.URL{Path: url}
	return u.String()
}

//goroutine
func Scrape(pages []conf.PageConfig, mode int) (ret []SearchResult) {
	c := conf.GetConf()
	resCh := make(chan SearchResult)

	var wg sync.WaitGroup
	// set all request first
	wg.Add(len(pages))

	//TODO: to deal with multiple keyworks

	// execute all request by goroutine
	for _, page := range pages {
		//TODO:interface
		switch mode {
		case 1:
			ind := indeed{page, c.Keywords[0].Search}
			go callScraper(&ind, resCh, &wg)
		case 2:
			stc := stackoverflow{page, c.Keywords[0].Search}
			go callScraper(&stc, resCh, &wg)
		case 3:
			stc := linkedin{page, c.Keywords[0].Search}
			go callScraper(&stc, resCh, &wg)
		default:
		}
		//go func(s Scraper, wg *sync.WaitGroup) {
		//	s.scrape(0, resCh, nil)
		//	wg.Done()
		//}(&ind, &wg)
	}

	// close channel when finishing all goroutine
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// wait until results channel is closed.
	for result := range resCh {
		ret = append(ret, result)
	}
	return
}
