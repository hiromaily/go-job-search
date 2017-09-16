package scrape

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	conf "github.com/hiromaily/go-job-search/libs/config"
	"github.com/hiromaily/go-job-search/libs/enum"
	lg "github.com/hiromaily/golibs/log"
	//"strconv"
	"net/http"
	"strings"
	"sync"
)

type linkedin struct {
	conf.PageConfig
	keyword string
}

var (
	urlSuffix   string = "%3A0"
	replaceData string = "& quot;"
)

// notify implements a method with a pointer receiver.
func (lkd *linkedin) scrape(start int, ret chan SearchResult, wg *sync.WaitGroup) {
	var waitGroup sync.WaitGroup

	//curl 'https://www.linkedin.com/jobs/search/?keywords=golang&location=Spain&locationId=es%3A0'
	url := fmt.Sprintf(lkd.Url+lkd.Param, lkd.keyword, encode(enum.COUNTRY[lkd.Country]), lkd.Country) + urlSuffix
	if start != 0 {
		url = fmt.Sprintf("%s&start=%d", url, start)
	}

	//https://www.linkedin.com/jobs/search/?keywords=golang&location=United%20Kingdom&locationId=gb%3A0&start=25
	lg.Debug("[URL]", url)

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

	//debug HTML
	//if url == "https://www.linkedin.com/jobs/search/?keywords=golang&location=German&locationId=de%3A0"{
	html, _ := doc.Find("body").Html()
	if html == "" {
		lg.Errorf("[scrape() for linkedin] no body: url:%s", url)
		if wg != nil {
			wg.Done()
		}
		return
	}
	//lg.Debugf("[scrape() for linkedin] url:%s\n%s", url, html)

	//get code
	doc.Find("code").Each(func(_ int, s *goquery.Selection) {
		res, _ := s.Html()
		lg.Debug("[HTML]", url, res)
		lg.Debug("------------------------")

		//convert to Json
		convertJson(&res)
	})

	//paging
	//&quot;paging&quot;:{&quot;total&quot;:157,&quot;count&quot;:25,&quot;start&quot;:25,&quot;links&quot;:[]}}
	//if start == 0 {
	//	searchCount := []int{}
	//	searchDoc := doc.Find("#searchCount").First()
	//	tmp := strings.Split(searchDoc.Text(), " ")
	//	for _, v := range tmp {
	//		if i, ok := strconv.Atoi(v); ok == nil {
	//			searchCount = append(searchCount, i)
	//		}
	//	}
	//	//lg.Debug("[searchCount]", searchCount)
	//
	//	// call left pages.
	//	if len(searchCount) == 3 {
	//		for i := 10; i < searchCount[2]; i += 10 {
	//			waitGroup.Add(1)
	//			go lkd.scrape(i, ret, &waitGroup)
	//		}
	//	}
	//}

	titles := SearchResult{Country: lkd.Country, BaseUrl: lkd.Url}
	jobs := []Job{}

	//analyze title
	doc.Find("h2.jobtitle a").Each(func(_ int, s *goquery.Selection) {
		//link
		link, _ := s.Attr("href")

		//company
		var company string

		//s.Parent().Next().Find("span").Each(func(_ int, ss *goquery.Selection) {
		companyDoc := s.Parent().Next().Find("span").First()
		tmpcom := strings.Trim(companyDoc.Text(), " \n")
		if tmpcom != "" {
			company = tmpcom
		}

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
	// -H 'Accept-Encoding: gzip, deflate, br'
	// -H 'Accept-Language: en-US,en;q=0.8,ja;q=0.6,nl;q=0.4,de;q=0.2,fr;q=0.2'
	// -H 'Upgrade-Insecure-Requests: 1'
	// -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36'
	// -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8'
	// -H 'Cache-Control: max-age=0'
	// -H 'Cookie: bcookie="v=2&ffc4d95d-f742-4342-82af-6ae5a6c9160b"; bscookie="v=1&20170709135329d8e98587-16be-42a4-8399-94f52289e29cAQEf8DB7AeoXVU5EdxlNFgK48cp-nI_x"; visit="v=1&M"; _chartbeat2=_a_R9D0yK8dDgBJW-.1501882194333.1501882194401.1; u_tz=GMT+0200; cap_session_id="1734705601:1"; PLAY_SESSION=65920ad176d27cc062587dc381fc8c1e32c177f6-chsInfo=7428b241-e4c7-4393-96a0-7fcc854ad3b8+premium_job_details_upsell_job_poster; sl="v=1&UCxQq"; JSESSIONID="ajax:3629013308086232428"; liap=true; li_at=AQEDARVE9GsDaTBEAAABXn_NrZoAAAFeo9oxmk4AzJANO4nxkkKQl5Jrk5fwdeadhUtqptWJBbW0Sy5I2zN4Zzh04ydK2j21LWapNyTCzPRCO93VeAAw6YYCwq1U1635Vgwaaj40foKfSjvE9W0qEm91; sdsc=22%3A1%2C1505546260670%7ECONN%2C02dVru8ipvlZ0%2B%2BfiGJokvVgHkAE%3D; lang="v=2&lang=en-us"; _gat=1; _ga=GA1.2.70720582.1499789675; _lipt=CwEAAAFei0cmtGvjsbgPdW-_MslEWBR5OBh-94J-DwQK9GcT_fvgoKEM622S-KCZ0j5kOoHidMYjTHirmLoRDCTpdTCrXG40i9_F2fwZIEJFxOu6I2KO5sM225gPZ3QSJIl0WlG6S1qEG2gbU3w3kRp9DpjNGZn0vRYEKPy9xq7L7IxGfVmx4Xbqr44lTEWVmM27rYW1ehfnHh75IU7aSJBcvlIP8FCBX-J7tW47zpaLDlezheddWHewM9QFedPEE5vp59C-WspNI7uZKybtUI6IrpqM07Qu94traf9OklR1YbfRUe2AgDUz-OddGUO7-zDLJO-KocU5amCjkb0nx4xbh0deuUbUtMSyCTGkclkdP9jnvqIOKM3IgXPobYu_vKZ3l06fAltClAwwcqqrpmirrpAPeGTQDKBqD3XTlg; lidc="b=VGST03:g=510:u=1:i=1505575256:t=1505661656:s=AQEuAl3Wq-RlJdxxh_HDWjqdVxozXGfu"'
	// -H 'Connection: keep-alive' --compressed
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,ja;q=0.6,nl;q=0.4,de;q=0.2,fr;q=0.2")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")

	//dump
	//dumpHTTP(req)

	// Send
	client := &http.Client{}
	return client.Do(req)
}

func convertJson(html *string) {
	list := [10][2]string{
		{"<!--", ""},
		{"-->", ""},
	}
	for _, data := range list {
		*html = strings.Replace(*html, data[0], data[1], -1)
	}

}
