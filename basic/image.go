package basic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/asciimoo/colly"
	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/model"
	"github.com/shohi/goinsight/util"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// JSONImageInsighter - fetch images by given json url
// TODO: use channel to async and decorate url with `#ID` instead of using global map
type JSONImageInsighter struct {
	Config config.JSONImageConfig
}

// ImageInsighter - fetch images, urls follow below form
// index page and detail & next page, final image url in detail page
type ImageInsighter struct{}

var logger = zap.NewExample().Sugar()

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
		util.Download(link, "", false)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	c.Visit(entryURL)
}

// Insight - insight image
func (i *JSONImageInsighter) Insight(ctx context.Context) {
	defer logger.Sync()

	// Instantiate default collector
	c := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	c.CacheDir = "./_cache"

	// Set URLs
	m := i.getImageURLs(i.Config.URL)

	// OnHTML must be set before Visit
	c.OnHTML("div.wp #container a[data-id] img[data-original]", func(e *colly.HTMLElement) {
		txn := config.DB.NewTransaction(true)
		defer txn.Discard()

		link := e.Attr("data-original")
		logger.Infow("", zap.String("link", link))
		fp := i.filepath(link, e.Request.Ctx.Get("ID"))
		err := util.Download(link, fp, false)
		val := "1"
		if err != nil {
			logger.Infow("failed to download image",
				"url", e.Request.URL.String(),
				"error", err.Error(),
			)
			val = "0"
		}

		err = txn.Set([]byte(link), []byte(val), byte(0))
		if err != nil {
			logger.Info(err.Error())
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

// NewJSONImageInsighter -- create new JSONImageInsighter using configuration
func NewJSONImageInsighter(v *viper.Viper) *JSONImageInsighter {
	var cfg config.JSONImageConfig
	err := v.Unmarshal(cfg)
	if err != nil {
		logger.Info(err)
		return nil
	}
	return &JSONImageInsighter{cfg}
}

func (i *JSONImageInsighter) getImageURLs(baseURL string) map[string]string {
	m := make(map[string]string)
	u, _ := url.Parse(baseURL)
	rootURL := u.Scheme + "://" + u.Host

	info, err := i.LoadImageJSON(baseURL)
	if err != nil {
		fmt.Println(err)
		return m
	}

	imageList := info.(*model.ImageCollection)
	for _, v := range imageList.List {
		click, err := strconv.Atoi(v.Click)
		if err != nil || click < i.Config.ThresHold {
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
			i.appendImageURLs(rootURL, u, m)
		}(url)
	}

	wg.Wait()
	return m
}

func (i *JSONImageInsighter) appendImageURLs(rootURL, jsonURL string, m map[string]string) {
	info, err := i.LoadImageJSON(jsonURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	//
	imageList := info.(*model.ImageCollection)
	for _, v := range imageList.List {
		click, err := strconv.Atoi(v.Click)
		if err != nil || click < i.Config.ThresHold {
			continue
		}
		tURL := rootURL + v.URL
		m[tURL] = v.ID
	}

	return
}

func (i *JSONImageInsighter) filepath(url, subdir string) (fp string) {
	filename := util.FilenameFromURL(url)
	if subdir == "" {
		fp = filepath.Join(i.Config.DirName, filename)
	} else {
		fp = filepath.Join(i.Config.DirName, subdir, filename)
	}
	return
}

// LoadImageJSON -- loads image json info
func (i *JSONImageInsighter) LoadImageJSON(url string) (interface{}, error) {
	body, err := util.GetContent(url)
	if err != nil {
		return nil, err
	}

	var info model.ImageCollection

	// The BOM identifies that the text is UTF-8 encoded, but it should be removed before decoding.
	// https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(body, &info)

	if err != nil {
		return nil, err
	}

	return &info, nil
}
