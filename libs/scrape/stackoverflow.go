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

type stackoverflow struct {
	conf.PageConfig
	keyword string
}

// notify implements a method with a pointer receiver.
func (sof *stackoverflow) scrape(start int, ret chan SearchResult, wg *sync.WaitGroup) {
	var waitGroup sync.WaitGroup

	///jobs/developer-jobs-using-go?pg=2
	//lg.Debug("[URL]", fmt.Sprintf("%s%s?pg=%d", sof.Url, sof.Param, start))
	url := sof.Url + sof.Param
	if start != 0 {
		url = fmt.Sprintf("%s?pg=%d", url, start)
	}
	doc, err := goquery.NewDocument(url)
	if err != nil {
		lg.Errorf("[scrape() for stackoverflow]")
		if wg != nil {
			wg.Done()
		}
		return
	}

	//debug HTML
	//res, _ := doc.Find("body").Html()
	//lg.Debug("[scrape for stackoverflow]", res)

	//paging
	if start == 0 {
		searchCount := []int{}
		linkDoc := doc.Find("div.pagination a.job-link").First()
		page, _ := linkDoc.Attr("title")
		tmp := strings.Split(page, " ")
		for _, v := range tmp {
			if i, ok := strconv.Atoi(v); ok == nil {
				searchCount = append(searchCount, i)
			}
		}

		//page 1 of 3
		//lg.Debug("[searchCount]", searchCount)

		// call left pages.
		if len(searchCount) == 2 {
			for i := 2; i <= searchCount[1]; i++ {
				waitGroup.Add(1)
				go sof.scrape(i, ret, &waitGroup)
			}
		}
	}

	jobs := []Job{}

	//analyze title
	doc.Find("h2.g-col10 a.job-link").Each(func(_ int, s *goquery.Selection) {
		//link
		link, _ := s.Attr("href")

		//company
		var company string
		companyDoc := s.Parent().Parent().Next().Find("div.-name").First()
		tmpcom := strings.Trim(companyDoc.Text(), " \n")
		if tmpcom != "" {
			company = tmpcom
		}

		//location
		var location string
		locationDoc := s.Parent().Parent().Next().Find("div.-location").First()
		tmploc := strings.Trim(locationDoc.Text(), " \n")
		tmploc = strings.Trim(tmploc, " -")
		tmploc = strings.Trim(tmploc, " \n")
		if tmploc != "" {
			location = tmploc
			//lg.Debug("location:", location)
		}

		if title, ok := s.Attr("title"); ok {
			//lg.Debug("title:", title)
			level := analyzeTitle(title)
			if level != 0 {
				jobs = append(jobs, Job{Title: title, Link: link, Company: company, City: location, MachingLevel: level})
			}
		}
	})

	//deliver by country
	if len(jobs) != 0 {
		sendJobs(jobs, sof.Url, ret)
	}

	//wait until all called
	if start == 0 {
		waitGroup.Wait()
	} else {
		wg.Done()
	}
}

func sendJobs(jobs []Job, url string, ret chan SearchResult) {
	country := "World"
	for i, value := range jobs {
		if value.City != "" {
			location := strings.Split(value.City, ",")

			jobs[i].City = strings.Trim(location[0], " ")
			location[1] = strings.Trim(location[1], " ")

			if len(location[1]) == 2 {
				if location[1] == "CA"{
					country = "Canada"
				}else{
					country = "USA"
				}
			} else {
				country = location[1]
			}
		}
		titles := SearchResult{Country: country, BaseUrl: url, Jobs: []Job{value}}
		ret <- titles
	}
}
