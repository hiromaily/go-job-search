package scrape

import (
	"net/http"
	ur "net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	conf "github.com/hiromaily/go-job-search/pkg/config"
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

func callScraper(s Scraper, resCh chan SearchResult, wg *sync.WaitGroup) {
	s.scrape(0, resCh, nil)
	wg.Done()
}

func analyzeTitle(title, keyword string) uint8 {
	//lg.Debug(title)
	switch keyword {
	case "golang":
		if strings.Index(title, "Golang") != -1 ||
			strings.Index(title, "Go ") != -1 ||
			strings.Index(title, "Go,") != -1 ||
			strings.Index(title, "Go)") != -1 {
			return 1
		}
	case "blockchain":
		if strings.Index(title, "Blockchain") != -1 ||
			strings.Index(title, "Block chain") != -1 ||
			strings.Index(title, "Block Chain") != -1 ||
			strings.Index(title, "Cryptography") != -1 {
			return 1
		}
	default:
		if strings.Index(title, keyword) != -1 {
			return 1
		}
	}

	return 0
}

//-----------------------------------------------------------------------------
// Utility function
//-----------------------------------------------------------------------------
func encode(url string) string {
	u := &ur.URL{Path: url}
	return u.String()
}

func getHTMLDocs(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(res.Body)
}

func sendRequest(url, cookies string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Set http header
	setHeader(req, cookies)

	//dump
	//dumpHTTP(req)

	// Send
	client := &http.Client{}
	return client.Do(req)
}

func setHeader(req *http.Request, cookies string) {
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,ja;q=0.6,nl;q=0.4,de;q=0.2,fr;q=0.2")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")

	//setLinkedinCookie(req)
	req.Header.Set("Cookie", cookies)
}
