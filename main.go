package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	colly "github.com/gocolly/colly/v2"
)

const URLFilePath = "./semi_url_list.txt"
const OutputCSVFile = "./output.csv"

type result struct {
	H1      string
	MetaDes string
	Title   string
	URL     string
}

var resultMap sync.Map

var channelH1 = make(chan result)
var channelMetaDes = make(chan result)
var channelTitle = make(chan result)

func main() {
	repository := repository{}

	go handleH1()
	go handleMetaDes()
	go handleTitle()

	URLs, err := repository.readLinesFromFile(URLFilePath)
	if err != nil {
		panic(err)
	}

	scrape(URLs)

	contents := make([][]string, 0, len(URLs))
	resultMap.Range(func(key, value interface{}) bool {
		row := make([]string, 0, 4)

		// H1
		if value.(result).H1 != "" {
			row = append(row, value.(result).H1)
		} else {
			row = append(row, "{no_h1}")
		}

		// meta_des
		if value.(result).MetaDes != "" {
			row = append(row, value.(result).MetaDes)
		} else {
			row = append(row, "{no_meta_des}")
		}

		// title
		if value.(result).Title != "" {
			row = append(row, value.(result).Title)
		} else {
			row = append(row, "{no_title}")
		}

		// URL
		row = append(row, value.(result).URL)

		contents = append(contents, row)

		return true
	})

	if err := repository.writeCSVToFile(OutputCSVFile, contents); err != nil {
		panic(err)
	}

	log.Println("Done.")
}

func scrape(URLs []string) {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.Async(),
		colly.IgnoreRobotsTxt(),
	)

	c.DisableCookies()

	c.WithTransport(&http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	})

	c.SetRequestTimeout(25 * time.Second)

	c.Limits([]*colly.LimitRule{
		{DomainGlob: "*altair.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*ansys.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*broadcom.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*cadence*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*dialog-semiconductor.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*siemens.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*globalfoundries.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*marvell.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*mediatek.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*novatek.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*nvidia.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*qualcomm.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*realtek.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*silvaco.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*synopsys.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*tsmc.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*umc.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
		{DomainGlob: "*xilinx.*", Parallelism: 5, Delay: 1 * time.Second, RandomDelay: 1 * time.Second},
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		channelH1 <- result{
			H1:  strings.TrimSpace(e.Text),
			URL: e.Request.URL.String(),
		}
	})
	c.OnHTML("meta", func(e *colly.HTMLElement) {
		if e.Attr("name") == "description" {
			channelMetaDes <- result{
				MetaDes: strings.TrimSpace(e.Attr("content")),
				URL:     e.Request.URL.String(),
			}
		}
	})
	c.OnHTML("title", func(e *colly.HTMLElement) {
		channelTitle <- result{
			Title: strings.TrimSpace(e.Text),
			URL:   e.Request.URL.String(),
		}
	})

	header := http.Header{}
	header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0")
	header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	header.Add("Accept-Language", "zh-TW,zh;q=0.8,en-US;q=0.5,en;q=0.3")
	header.Add("Cache-Control", "max-age=0")
	header.Add("Upgrade-Insecure-Requests", "1")
	header.Add("Connection", "keep-alive")

	c.OnError(func(r *colly.Response, e error) {
		thisURL := r.Request.URL.String()

		if r.Request.URL.Host == "" ||
			(r.StatusCode > 399 && r.StatusCode < 408) ||
			(r.StatusCode > 408 && r.StatusCode < 600) {

			skipMsg := fmt.Sprintf("{skip_%d}", r.StatusCode)

			resultMap.Store(thisURL, result{
				H1:      skipMsg,
				MetaDes: skipMsg,
				Title:   skipMsg,
				URL:     thisURL,
			})

			return
		}

		if strings.Contains(e.Error(), "no such host") {
			skipMsg := "skip_no_such_host"

			resultMap.Store(thisURL, result{
				H1:      skipMsg,
				MetaDes: skipMsg,
				Title:   skipMsg,
				URL:     thisURL,
			})

			return
		}

		log.Printf("Error on %s ||| %+v", thisURL, e)

		time.Sleep(3 * time.Second)

		c.Request("GET", thisURL, nil, nil, header) // retry
	})

	log.Printf("Start with %d URLs\n", len(URLs))
	for _, URL := range URLs {
		c.Request("GET", URL, nil, nil, header)
	}

	c.Wait()
}

func handleH1() {
	for {
		signal := <-channelH1
		resultOfThisURL, exists := resultMap.Load(signal.URL)
		if exists {
			// UPDATE
			toUpdate := resultOfThisURL.(result)

			if toUpdate.H1 != "" { // we don't need second H1 field
				continue
			}

			toUpdate.H1 = signal.H1
			resultMap.Store(signal.URL, toUpdate)
		} else {
			// INSERT
			resultMap.Store(signal.URL, signal)
		}
	}
}

func handleMetaDes() {
	for {
		signal := <-channelMetaDes
		resultOfThisURL, exists := resultMap.Load(signal.URL)
		if exists {
			// UPDATE
			toUpdate := resultOfThisURL.(result)

			if toUpdate.MetaDes != "" { // we don't need second meta_des field
				continue
			}

			toUpdate.MetaDes = signal.MetaDes
			resultMap.Store(signal.URL, toUpdate)
		} else {
			// INSERT
			resultMap.Store(signal.URL, signal)
		}
	}
}

func handleTitle() {
	for {
		signal := <-channelTitle
		resultOfThisURL, exists := resultMap.Load(signal.URL)
		if exists {
			// UPDATE
			toUpdate := resultOfThisURL.(result)

			if toUpdate.Title != "" { // we don't need second title field
				continue
			}

			toUpdate.Title = signal.Title
			resultMap.Store(signal.URL, toUpdate)
		} else {
			// INSERT
			resultMap.Store(signal.URL, signal)
		}
	}
}
