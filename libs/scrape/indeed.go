package scrape

import (
	"github.com/PuerkitoBio/goquery"
	conf "github.com/hiromaily/go-job-search/libs/config"
	lg "github.com/hiromaily/golibs/log"
	"sync"
)

func scrapeIndeed(url string, ret chan []string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		lg.Errorf("[scrapeIndeed]")
		return
	}

	doc.Find("h2.jobtitle a").Each(func(_ int, s *goquery.Selection) {
		//if jsonData, ok := s.Attr("id"); ok {
		//
		//	//decode
		//	htmlStringDecode(&jsonData)
		//
		//	//analyze json object
		//	var jsonObject map[string]interface{}
		//	//json.JsonAnalyze(jsonData, &jsonObject)
		//	json.Unmarshal([]byte(jsonData), &jsonObject)
		//
		//	//extract date from json object
		//	//e.g. 2016-02-27 03:30:00
		//
		//	strDate := jsonObject["field19"].(string)
		//	if isTimeApplicable(strDate) {
		//		dates = append(dates, strDate)
		//	}
		//}
	})

	ret <- []string{"test1", "test2"}
}

//func perseHTML(htmldata *goquery.Document) []string {
//}

//goroutine
func ScrapeIndeed(confs []conf.PageIndeedConfig) (ret []string) {
	resCh := make(chan []string)

	var waitGroup sync.WaitGroup
	// set all request first
	waitGroup.Add(len(confs))

	// execute all request by goroutine
	for _, conf := range confs {
		go func(u string) {
			scrapeIndeed(u, resCh)
			waitGroup.Done()
		}(conf.Url)
	}

	// close channel when finishing all goroutine
	go func() {
		waitGroup.Wait()
		close(resCh)
	}()

	// wait until results channel is closed.
	for result := range resCh {
		lg.Infof("[ScrapeIndeed] result: %v\n", result)
		ret = append(ret, result...)
	}
	return
}
