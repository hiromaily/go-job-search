package scrape

import (
	"fmt"
	"html"
	"log"
	"sync"

	"github.com/PuerkitoBio/goquery"
	jp "github.com/buger/jsonparser"
	conf "github.com/hiromaily/go-job-search/pkg/config"
	"github.com/hiromaily/go-job-search/pkg/enum"
	lg "github.com/hiromaily/golibs/log"
	ck "github.com/hiromaily/golibs/web/cookie"
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
	urlSuffix      string = "%3A0"
	linkedinCookie string
)

func init() {
	err := getLinkedInCookie()
	if err != nil {
		log.Println(err)
	}
}

func getLinkedInCookie() error {
	var (
		//url = "www.linkedin.com"
		url = "linkedin.com"
	)

	if linkedinCookie == "" {
		cookies, err := ck.GetAllValue(url)
		if err != nil {
			return err
		}
		for key, value := range cookies {
			linkedinCookie = linkedinCookie + fmt.Sprintf("%s=\"%s\"; ", key, value)
		}
	}
	return nil
}

// notify implements a method with a pointer receiver.
func (lkd *linkedin) scrape(start int, ret chan SearchResult, wg *sync.WaitGroup) {
	var waitGroup sync.WaitGroup

	// create URL
	url := fmt.Sprintf(lkd.Url+lkd.Param, lkd.keyword, encode(enum.CountryMaps[lkd.Country]), lkd.Country) + urlSuffix
	if start != 0 {
		url = fmt.Sprintf("%s&start=%d", url, start)
	}

	// get body
	doc, err := getHTMLDocsWithCookie(url, linkedinCookie)
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

		data := analyzeJson([]byte(code))
		if data != nil {
			//lg.Debug("[analyzeJson]", data)
			if result, ok = data.(linkedinResult); !ok {
				//end
				lg.Errorf("[scrape() for linkedin] data can no be fetched from code elements: url:%s", url)
			}
		}
	})

	// paging, call all existing pages by paging information
	if start == 0 {

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

func analyzeJson(data []byte) interface{} {
	//job info
	_, _, _, err := jp.Get(data, "data", "metadata", "locationInfo")
	if err == nil {
		result := linkedinResult{}

		//jobs := make([]map[string]string)
		jobs := []map[string]string{}
		jp.ArrayEach(data, func(value []byte, dataType jp.ValueType, offset int, err error) {

			//check companyDetails
			_, _, _, err = jp.Get(value, "companyDetails")
			if err != nil {
				return
			}

			//job title
			title, err := jp.GetString(value, "title")

			//job link
			link, err := jp.GetString(value, "applyMethod", "companyApplyUrl")

			//job location
			location, err := jp.GetString(value, "formattedLocation")

			//TODO: how to get company name...
			jobs = append(jobs, map[string]string{
				"location": location,
				"title":    title,
				"company":  "",
				"link":     link,
			})

			//link
			//link, err := jp.GetString(value, "hitInfo", "com.linkedin.voyager.search.SearchJobJserp", "jobPosting")
			//if err == nil {
			//	tmp := strings.Split(link, ":")
			//	link = tmp[len(tmp)-1]
			//}
			//title, company, location
			//job, _, _, err := jp.Get(value, "hitInfo", "com.linkedin.voyager.search.SearchJobJserp", "jobPostingResolutionResult")
			//if err == nil {
			//	formattedLocation, _ := jp.GetString(job, "formattedLocation")
			//	title, _ := jp.GetString(job, "title")
			//	company, _ := jp.GetString(job, "companyDetails", "com.linkedin.voyager.jobs.JobPostingCompany", "companyResolutionResult", "name")
			//	jobs = append(jobs, map[string]string{
			//		"location": formattedLocation,
			//		"title":    title,
			//		"company":  company,
			//		"link":     fmt.Sprintf("/jobs/view/%s/", link),
			//	})
			//}

		}, "included")
		result.jobs = jobs

		//count
		result.count, err = jp.GetInt(data, "data", "paging", "count")
		if err == nil {
			result.start, _ = jp.GetInt(data, "data", "paging", "start")
			result.total, _ = jp.GetInt(data, "data", "paging", "total")
		}
		//lg.Debug(result.count)
		//lg.Debug(result.start)
		//lg.Debug(result.total)

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
