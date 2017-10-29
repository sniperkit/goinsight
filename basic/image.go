package basic

import (
	"fmt"
	"net/url"

	"github.com/asciimoo/colly"
	"github.com/shohi/goinsight/model"
	"github.com/shohi/goinsight/util"
)

// InsightImage - fetch images, urls follow below form
// index page and detail & next page, final image url in detail page
func InsightImage(entryURL string) {

	// Instantiate default collector
	c := colly.NewCollector()
	detailCollector := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	c.CacheDir = "./image_cache"
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

// InsightJSONImage - fetch images by given json url
func InsightJSONImage(jsonURL string) {

	// Instantiate default collector
	c := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	// c.CacheDir = "./image_cache"

	// Set URLs
	m := make(map[string]string)
	u, _ := url.Parse(jsonURL)
	rootURL := u.Scheme + "://" + u.Host
	info, err := util.LoadImageJSON(jsonURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	imageList := info.(*model.ImageCollection)
	for _, v := range imageList.List {
		tURL := rootURL + v.URL
		m[tURL] = v.ID
	}

	// OnHTML must set before visit
	c.OnHTML("div.wp #container a[data-id] img[data-original]", func(e *colly.HTMLElement) {
		link := e.Attr("data-original")
		fmt.Println(link)
		util.Download(link, e.Request.Ctx.Get("ID"))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("ID", m[r.URL.String()])
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scrapping
	for url := range m {
		c.Visit(url)
	}
}

func getImageURLs(baseURL string) (m map[string]string) {
	u, _ := url.Parse(baseURL)
	rootURL := u.Scheme + "://" + u.Host

	info, err := util.LoadImageJSON(baseURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	imageList := info.(*model.ImageCollection)
	for _, v := range imageList.List {
		tURL := rootURL + v.URL
		m[tURL] = v.ID
	}

	// load other url
	jsonURLs := util.GetPageURLs(baseURL, imageList.Pages, true)
	if len(jsonURLs) < 1 {
		return
	}

	for _, url := range jsonURLs {
		appendImageURLs(rootURL, url, m)
	}

	return
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
		tURL := rootURL + v.URL
		m[tURL] = v.ID
	}

	return
}
