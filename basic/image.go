package basic

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/asciimoo/colly"
	"github.com/shohi/goinsight/model"
	"github.com/shohi/goinsight/util"
)

// JSONImageInsighter - fetch images by given json url
// TODO: use channel to async and decorate url with `#ID` instead of using global map
type JSONImageInsighter struct{}

// ImageInsighter - fetch images, urls follow below form
// index page and detail & next page, final image url in detail page
type ImageInsighter struct{}

var clickThreshold = 50000

// Insight - insight image
func (i *ImageInsighter) Insight(entryURL string) {

	// Instantiate default collector
	c := colly.NewCollector()
	detailCollector := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	c.CacheDir = "./_cache"
	detailCollector.CacheDir = c.CacheDir

	// On every a element which has href attribute call callback
	c.OnHTML("div.content.masonry.on div.mbitem div.mbpic.mbpic2 a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println("link ==> ", link)
		// Visit link found on page and download images
		detailCollector.Visit(e.Request.AbsoluteURL(link))
	})

	detailCollector.OnHTML("div.wp #container a[data-id] img[data-original]", func(e *colly.HTMLElement) {
		link := e.Attr("data-original")
		util.Download(link, "")
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	c.Visit(entryURL)
}

// Insight - insight image
func (i *JSONImageInsighter) Insight(jsonURL string) {

	// Instantiate default collector
	c := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	c.CacheDir = "./_cache"

	// Set URLs
	m := getImageURLs(jsonURL)

	// OnHTML must be set before Visit
	c.OnHTML("div.wp #container a[data-id] img[data-original]", func(e *colly.HTMLElement) {
		link := e.Attr("data-original")
		fmt.Println(link)
		if err := util.Download(link, e.Request.Ctx.Get("ID")); err != nil {
			log.Println(err.Error())
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("ID", m[r.URL.String()])
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scrapping
	for url := range m {
		go func(u string) {
			c.Visit(u)
		}(url)
	}
	c.Wait()
}

func getImageURLs(baseURL string) map[string]string {
	m := make(map[string]string)
	u, _ := url.Parse(baseURL)
	rootURL := u.Scheme + "://" + u.Host

	info, err := util.LoadImageJSON(baseURL)
	if err != nil {
		fmt.Println(err)
		return m
	}

	imageList := info.(*model.ImageCollection)
	for _, v := range imageList.List {
		click, err := strconv.Atoi(v.Click)
		if err != nil || click < clickThreshold {
			continue
		}
		tURL := rootURL + v.URL
		m[tURL] = v.ID
	}

	// load other url
	jsonURLs := util.GetPageURLs(baseURL, imageList.Pages, true)
	if len(jsonURLs) < 1 {
		return m
	}

	// use goroutines to load json url
	var wg sync.WaitGroup
	wg.Add(len(jsonURLs))

	for _, url := range jsonURLs {
		go func(u string) {
			defer wg.Done()
			appendImageURLs(rootURL, u, m)
		}(url)
	}

	wg.Wait()
	return m
}

func appendImageURLs(rootURL, jsonURL string, m map[string]string) {
	info, err := util.LoadImageJSON(jsonURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	//
	imageList := info.(*model.ImageCollection)
	for _, v := range imageList.List {
		click, err := strconv.Atoi(v.Click)
		if err != nil || click < clickThreshold {
			continue
		}
		tURL := rootURL + v.URL
		m[tURL] = v.ID
	}

	return
}
