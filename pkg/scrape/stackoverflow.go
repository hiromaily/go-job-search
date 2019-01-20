package scrape

import (
	"fmt"
	"github.com/bookerzzz/grok"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	conf "github.com/hiromaily/go-job-search/pkg/config"
	lg "github.com/hiromaily/golibs/log"
	ck "github.com/hiromaily/golibs/web/cookie"
)

type stackoverflow struct {
	conf.PageConfig
	keyword string
}

var (
	stackoverflowCookie string
)

func init() {
	getStackOverFlowCookie()
}

func getStackOverFlowCookie() {
	var (
		url = "stackoverflow.com"
	)

	if stackoverflowCookie == "" {
		cookies := ck.GetAllValue(url)
		for key, value := range cookies {
			stackoverflowCookie = stackoverflowCookie + fmt.Sprintf("%s=\"%s\"; ", key, value)
		}
	}
}

// notify implements a method with a pointer receiver.
func (sof *stackoverflow) scrape(start int, ret chan SearchResult, wg *sync.WaitGroup) {
	var waitGroup sync.WaitGroup

	///jobs/developer-jobs-using-go?pg=2
	url := sof.Url + sof.Param
	if start != 0 {
		url = fmt.Sprintf("%s?pg=%d", url, start)
	}

	// get body
	//doc, err := getHTMLDocs(url)
	doc, err := getHTMLDocsWithCookie(url, stackoverflowCookie)
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
	doc.Find(".-job-summary").Each(func(_ int, s *goquery.Selection) {

		//title object
		titleDoc := s.Find(".-title h2.job-details__spaced a").First()

		//title
		title := titleDoc.Text()
		//lg.Debug(title)

		//link
		link, _ := titleDoc.Attr("href")
		//lg.Debug(link)

		//company
		companyDoc := s.Find(".-company span")
		company := companyDoc.First().Text()
		company = strings.Trim(company, " \n")
		//lg.Debug(company)

		//location
		location := companyDoc.First().Next().Text()
		if len(strings.Split(location, "-")) == 2 {
			location = strings.Split(location, "-")[1]
		}
		location = strings.Trim(location, " \n")
		//lg.Debug(location)

		level := analyzeTitle(title, sof.keyword)
		if level != 0 {
			jobs = append(jobs, Job{Title: title, Link: link, Company: company, City: location, MachingLevel: level})
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
	grok.Value(jobs)

	//format country name
	country := "World"
	for i, job := range jobs {
		country = "World"

		if job.City != "" && strings.Index(job.City, "No office location") == -1 {

			location := strings.Split(job.City, ",")

			//Attention
			//sometimes location is just `No office location` and len(location) = 1
			if len(location) == 1 {
				country = location[0]
			} else {
				jobs[i].City = strings.Trim(location[0], " ")

				location[1] = strings.Trim(location[1], " ")

				//in case of 2 letters
				if len(location[1]) == 2 && location[1] != "UK" {
					if location[1] == "CA" {
						country = "Canada"
					} else {
						country = "USA"
					}
				} else {
					country = location[1]
				}
				// rename country
				if country == "Deutschland" {
					country = "Germany"
				} else if country == "Czechia" {
					country = "Czech"
				}
			}
		}

		//send
		titles := SearchResult{Country: country, BaseUrl: url, Jobs: []Job{job}}
		ret <- titles
	}
}
