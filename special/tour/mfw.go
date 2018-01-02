package tour

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/asciimoo/colly"
	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/util"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var logger = zap.NewExample().Sugar()

// MfwTourInsighter ...
type MfwTourInsighter struct {
	Config config.MfwImageConfig

	pageURLs []string
	client   fasthttp.Client
}

// NewMfwTourInsighter -- create new MfwTourInsighter using configuration
func NewMfwTourInsighter(v *viper.Viper) *MfwTourInsighter {
	var cfg config.MfwImageConfig

	err := v.Unmarshal(&cfg.CommonConfig)
	if err != nil {
		logger.Info(err)
		return nil
	}

	logger.Info(cfg)
	return &MfwTourInsighter{Config: cfg}
}

// Insight - insight image, ref http://blog.csdn.net/qijingpei/article/details/77668972
// Implement interface
func (s *MfwTourInsighter) Insight(ctx context.Context) {
	defer logger.Sync()

	// Instantiate default collector
	c := colly.NewCollector()
	detailCollector := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	c.CacheDir = s.Config.CacheDir
	detailCollector.CacheDir = s.Config.CacheDir

	if s.Config.NewCache {
		os.RemoveAll(c.CacheDir)
	}

	if s.Config.NewDownload {
		os.RemoveAll(s.Config.DownloadDir)
	}

	// 1. Get URLs
	if err := s.getPageURLs(); err != nil {
		logger.Infow("get total pages error", "error", err)
		return
	}
	logger.Infow("total pages", "count", len(s.pageURLs))

	// OnHTML must be set before Visit. On each page list
	c.OnHTML("div.post-list li div.post-cover a", func(e *colly.HTMLElement) {
		txn := config.DB.NewTransaction(true)
		defer txn.Discard()

		link := e.Attr("href")
		logger.Infow("", zap.String("link", link))
		// Check whether or not in database
		item, err := txn.Get([]byte(link))
		if err == nil {
			_, e := item.Value()
			if e == nil {
				return
			}
		}
		logger.Infow("", zap.String("link", link))
		u, err := url.Parse(e.Request.URL.String())
		detailLink := u.Scheme + "://" + u.Host + link

		err = detailCollector.Visit(detailLink)
		if err != nil {
			logger.Infow("detail fetching error", "error", err)
		}
		err = txn.Set([]byte(link), []byte("0"), byte(0))
		if err != nil {
			logger.Info(err.Error())
		}
	})

	// on each page, fetch images
	detailCollector.OnHTML("div.vc_article div._j_content_box div.add_pic a", func(e *colly.HTMLElement) {
		baseDir := e.Request.Ctx.Get("BaseDir")
		if baseDir == "" {
			return
		}

		href := e.Attr("href")
		filename, err := util.GetResourceName(href)
		if err != nil {
			return
		}

		link, ok := e.DOM.Find("img").First().Attr("data-rt-src")
		if !ok {
			return
		}

		suffix, err := util.GetResourceSuffix(link)
		if err != nil {
			return
		}

		fp := filepath.Join(s.Config.DownloadDir, baseDir, filename+suffix)
		err = util.Download(link, fp, false)
		if err != nil {
			logger.Infow("failed to download image",
				"url", e.Request.URL.String(),
				"error", err.Error(),
			)
		}
	})

	// Before making a request print "Visiting ..."
	detailCollector.OnRequest(func(r *colly.Request) {
		logger.Infow("requests", "url", r.URL.String())
		baseDir, err := util.GetResourceName(r.URL.String())
		if err == nil {
			r.Ctx.Put("BaseDir", baseDir)
		} else {
			logger.Infow("get resource name error", "error", err)
		}
	})

	// Start scrapping
	for _, url := range s.pageURLs {
		go func(u string) {
			logger.Info(u)
			c.Visit(u)
		}(url)
	}
	c.Wait()
}

func (s *MfwTourInsighter) getPageURLs() error {
	homePage := fmt.Sprintf(s.Config.URL, 1)

	statusCode, body, err := s.client.Get(nil, homePage)
	logger.Infow("", "home_page", homePage)
	if err != nil {
		return err
	}

	if statusCode != fasthttp.StatusOK {
		return errors.New("Status Code Is Not OK")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	numStr := doc.Find("div._pagebar div span.count span").First().Text()
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return err
	}

	if num > 0 {
		for k := 0; k < num; k++ {
			requestURL := fmt.Sprintf(s.Config.URL, k+1)
			s.pageURLs = append(s.pageURLs, requestURL)
		}
	}

	return nil
}
