package main

import (
	"log"
	"strings"
	"sync"

	colly "github.com/gocolly/colly/v2"
)

const URLFilePath = "./semi_url_list.txt"
const OutputCSVFile = "./output.csv"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0"

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

		if value.(result).H1 != "" {
			row = append(row, value.(result).H1)
		} else {
			row = append(row, "no_h1")
		}

		if value.(result).MetaDes != "" {
			row = append(row, value.(result).MetaDes)
		} else {
			row = append(row, "no_meta_des")
		}

		if value.(result).Title != "" {
			row = append(row, value.(result).Title)
		} else {
			row = append(row, "no_title")
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
	c := colly.NewCollector()
	c.UserAgent = userAgent
	c.Async = true

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

	c.OnRequest(func(r *colly.Request) {
		log.Println("Request: " + r.URL.String())
	})
	c.OnScraped(func(r *colly.Response) {
		log.Println("Scraped: " + r.Request.URL.String())
	})

	for _, URL := range URLs {
		c.Visit(URL)
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
