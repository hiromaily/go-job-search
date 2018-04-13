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
	var waitGroup sync.WaitGroup

	// http request
	url := fmt.Sprintf(ind.Url+ind.Param, ind.keyword, start)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		lg.Errorf("[scrape() for indeed] %s", url)
		if wg != nil {
			wg.Done()
		}
		return
	}

	//debug HTML
	//if url == "https://www.indeed.nl/vacatures?q=golang&start=0"{
	//	res, _ := doc.Find("body").Html()
	//	lg.Debug("[scrapeIndeed]", res)
	//}

	// paging, call all existing pages by paging information
	if start == 0 {
		searchCount := []int{}
		searchDoc := doc.Find("#searchCount").First()
		tmp := strings.Split(searchDoc.Text(), " ")
		for _, v := range tmp {
			if i, ok := strconv.Atoi(v); ok == nil {
				searchCount = append(searchCount, i)
			}
		}
		//lg.Debug("[searchCount]", searchCount)

		// call left pages.
		if len(searchCount) == 3 {
			for i := 10; i < searchCount[2]; i += 10 {
				waitGroup.Add(1)
				go ind.scrape(i, ret, &waitGroup)
			}
		}
	}

	// parse html
	titles := SearchResult{Country: ind.Country, BaseUrl: ind.Url}
	jobs := []Job{}

	//analyze title
	doc.Find("h2.jobtitle a").Each(func(_ int, s *goquery.Selection) {
		//link
		link, _ := s.Attr("href")

		//company
		var company string

		//this emement may be changeble
		companyDoc := s.Parent().Next()
		//companyDoc := s.Parent().Next().Find("span")
		//companyDoc := s.Parent().Next().Find("span").First()

		tmpcom := strings.Trim(companyDoc.Text(), " \n")
		if tmpcom != "" {
			company = tmpcom
		}

		if title, ok := s.Attr("title"); ok {
			level := analyzeTitle(title, ind.keyword)
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
