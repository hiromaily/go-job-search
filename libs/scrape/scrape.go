package scrape

import (
	conf "github.com/hiromaily/go-job-search/libs/config"
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

//goroutine
func Scrape(mode int) (ret []SearchResult) {
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
		ret = append(ret, result)
	}
	return
}
