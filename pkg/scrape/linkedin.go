package scrape

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	jp "github.com/buger/jsonparser"
	conf "github.com/hiromaily/go-job-search/pkg/config"
	"github.com/hiromaily/go-job-search/pkg/enum"
	lg "github.com/hiromaily/golibs/log"
	ck "github.com/hiromaily/golibs/web/cookie"
	"html"
	"net/http"
	"strings"
	"sync"
)

//TODO:login is required to fetch data.

type linkedin struct {
	conf.PageConfig
	keyword string
}

type linkedinResult struct {
	jobs  []map[string]string
	count int64
	start int64
	total int64
}

var (
	urlSuffix string = "%3A0"
	cookieVal string = ""
)

func init() {
	//get cookie from chrome
	getLinkedinCookie()
}

func getLinkedinCookie() {
	var (
		//url = "www.linkedin.com"
		url = "linkedin.com"
	)

	if cookieVal == "" {
		cookies := ck.GetAllValue(url)
		for key, value := range cookies {
			cookieVal = cookieVal + fmt.Sprintf("%s=\"%s\"; ", key, value)
		}
	}
}

// notify implements a method with a pointer receiver.
func (lkd *linkedin) scrape(start int, ret chan SearchResult, wg *sync.WaitGroup) {
	var waitGroup sync.WaitGroup

	// create URL
	url := fmt.Sprintf(lkd.Url+lkd.Param, lkd.keyword, encode(enum.COUNTRY[lkd.Country]), lkd.Country) + urlSuffix
	if start != 0 {
		url = fmt.Sprintf("%s&start=%d", url, start)
	}
	//https://www.linkedin.com/jobs/search/?keywords=golang&location=Netherlands&locationId=nl%3A0
	//https://www.linkedin.com/jobs/search/?keywords=golang&location=Netherlands&locationId=nl%3A0&start=25
	//lg.Debug("[URL]", url)

	// http request
	resp, err := sendRequest(url)
	var doc *goquery.Document
	if err == nil {
		//doc, err := goquery.NewDocument(url)
		doc, err = goquery.NewDocumentFromResponse(resp)
	}
	if err != nil {
		lg.Errorf("[scrape() for linkedin] %v", err)
		if wg != nil {
			wg.Done()
		}
		return
	}

	// check body
	body := doc.Find("body")
	if len(body.Nodes) == 0 {
		lg.Errorf("[scrape() for linkedin] no body: url:%s", url)
		if wg != nil {
			wg.Done()
		}
		return
	}

	// get <code> elements that includes any information
	var result linkedinResult
	var ok bool
	doc.Find("code").Each(func(_ int, s *goquery.Selection) {
		code, _ := s.Html()
		// unescape json data
		code = html.UnescapeString(code)

		//lg.Debugf("[HTML] url=%s\n%s", url, code)
		//lg.Debug("------------------------")

		data := analyzeJson([]byte(code))
		if data != nil {
			//lg.Debug("[analyzeJson]", data)
			if result, ok = data.(linkedinResult); !ok {
				//end
				lg.Errorf("[scrape() for linkedin] data can no be fetched from code elements: url:%s", url)
			}
		}
	})
	//if len(result.jobs) == 0{
	//	if wg != nil {
	//		wg.Done()
	//	}
	//	return
	//}

	// paging, call all existing pages by paging information
	if start == 0 {
		//paging
		//{"total":71,"count":25,"start":0,"links":[]}}
		//{"total":71,"count":25,"start":25,"links":[]}}

		// call left pages.
		if result.total > result.count {
			for i := result.count; i < result.total; i += result.count {
				waitGroup.Add(1)
				go lkd.scrape(int(i), ret, &waitGroup)
			}
		}
	}

	titles := SearchResult{Country: lkd.Country, BaseUrl: lkd.Url}
	jobs := []Job{}

	//analyze title
	for _, job := range result.jobs {
		//title
		if job["title"] != "" {
			level := analyzeTitle(job["title"], lkd.keyword)
			if level != 0 {
				jobs = append(jobs, Job{Title: job["title"], Link: job["link"], Company: job["company"], City: job["location"], MachingLevel: level})
			}
		}

	}
	titles.Jobs = jobs
	ret <- titles

	//wait until all called
	if start == 0 {
		waitGroup.Wait()
	} else {
		wg.Done()
	}
}

func sendRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Set http header
	setHeader(req)

	//dump
	//dumpHTTP(req)

	// Send
	client := &http.Client{}
	return client.Do(req)
}

func setHeader(req *http.Request) {
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,ja;q=0.6,nl;q=0.4,de;q=0.2,fr;q=0.2")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")

	//setLinkedinCookie(req)
	req.Header.Set("Cookie", cookieVal)

}

func analyzeJson(data []byte) interface{} {
	//job info
	_, _, _, err := jp.Get(data, "metadata", "locationInfo")
	if err == nil {
		result := linkedinResult{}

		//jobs := make([]map[string]string)
		jobs := []map[string]string{}
		jp.ArrayEach(data, func(value []byte, dataType jp.ValueType, offset int, err error) {
			//link
			link, err := jp.GetString(value, "hitInfo", "com.linkedin.voyager.search.SearchJobJserp", "jobPosting")
			if err == nil {
				tmp := strings.Split(link, ":")
				link = tmp[len(tmp)-1]
			}
			//title, company, location
			job, _, _, err := jp.Get(value, "hitInfo", "com.linkedin.voyager.search.SearchJobJserp", "jobPostingResolutionResult")
			if err == nil {
				formattedLocation, _ := jp.GetString(job, "formattedLocation")
				title, _ := jp.GetString(job, "title")
				company, _ := jp.GetString(job, "companyDetails", "com.linkedin.voyager.jobs.JobPostingCompany", "companyResolutionResult", "name")
				jobs = append(jobs, map[string]string{
					"location": formattedLocation,
					"title":    title,
					"company":  company,
					"link":     fmt.Sprintf("/jobs/view/%s/", link),
				})
			}

		}, "elements")
		result.jobs = jobs
		//count
		result.count, err = jp.GetInt(data, "paging", "count")
		if err == nil {
			result.start, _ = jp.GetInt(data, "paging", "start")
			result.total, _ = jp.GetInt(data, "paging", "total")
		}

		return result
	}

	return nil
}

//func convertJson(code *string) (*map[string]interface{}, error) {
//	var res map[string]interface{}
//	err := json.Unmarshal([]byte(*code), &res)
//	if err != nil {
//		return nil, err
//	}
//	//lg.Debug(res)
//	return &res, nil
//}
