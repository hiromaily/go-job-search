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

type indeed struct {
	conf.PageConfig
	keyword string
}

// notify implements a method with a pointer receiver.
func (ind *indeed) scrape(start int, ret chan SearchResult, wg *sync.WaitGroup) {
	//lg.Debug("[URL]", fmt.Sprintf(url+param, keyword,start))

	doc, err := goquery.NewDocument(fmt.Sprintf(ind.Url+ind.Param, ind.keyword, start))
	if err != nil {
		lg.Errorf("[scrapeIndeed]")
		return
	}

	//debug HTML
	//if url == "https://www.indeed.nl/vacatures?q=golang&start=0"{
	//	res, _ := doc.Find("body").Html()
	//	lg.Debug("[scrapeIndeed]", res)
	//}

	titles := SearchResult{Country: ind.Country, BaseUrl: ind.Url}
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
				go ind.scrape(i, ret, &waitGroup)
			}
		}
	}

	//analyze title
	doc.Find("h2.jobtitle a").Each(func(_ int, s *goquery.Selection) {
		//link
		link, _ := s.Attr("href")

		//company
		var company string
		s.Parent().Next().Find("span").Each(func(_ int, ss *goquery.Selection) {
			if strings.Trim(ss.Text(), " ") != "" {
				company = strings.Trim(ss.Text(), " \n")
			}
		})

		if title, ok := s.Attr("title"); ok {
			level := analyzeTitle(title)
			if level != 0 {
				jobs = append(jobs, Job{Title: title, Link: link, Company: company, MachingLevel: level})
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
