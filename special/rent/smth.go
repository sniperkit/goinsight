package rent

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"

	"github.com/PuerkitoBio/goquery"
	"github.com/asciimoo/colly"
	"github.com/deckarep/golang-set"
	"github.com/gocarina/gocsv"
	"github.com/jinzhu/now"

	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/util"
	"go.uber.org/zap"

	"github.com/tealeg/xlsx"
)

// SmthData ...
type SmthData struct {
	Title    string    `csv:"title"`
	Href     string    `csv:"href"`
	Author   string    `csv:"author"`
	Comments int       `csv:"comments"`
	Last     time.Time `csv:"time"`
}

// const baseURL = "http://www.newsmth.net/nForum/board/HouseRent?ajax"
// var u, _ = url.Parse(baseURL)
// var domainURL = u.Scheme + "://" + u.Host

func parseTime(timestr string) (t time.Time, err error) {
	if strings.Contains(timestr, "-") {
		t, err = now.Parse(timestr)

	} else if strings.Contains(timestr, ":") {
		t, err = now.Parse(timestr)

	}
	return
}

func (d *SmthData) populate(s *goquery.Selection, domainURL string) {
	d.Last = time.Now()
	d.Comments = 0

	// article info
	article := s.Find("td.title_9 a").First()

	href, exists := article.Attr("href")
	if exists {
		d.Href = domainURL + href
	}

	d.Title = article.Text()

	// user info
	user := s.Find("td.title_12 a").First()
	d.Author = user.Text()

	// comments info
	comment := s.Find("td.title_11").Get(2)
	commentNo, err := strconv.Atoi(comment.Data)
	if err != nil {
		d.Comments = commentNo
	}

	// last info
	last := s.Find("td.title_10 a").First()
	lasttime := strings.TrimSpace(last.Text())
	d.Last, err = parseTime(lasttime)
}

var logger = zap.NewExample().Sugar()

// NewSmthRentInsighter -- create new SmthRentInsighter using configuration
func NewSmthRentInsighter(v *viper.Viper) *SmthRentInsighter {
	var cfg config.SmthRentConfig

	// unmarshal direct fields
	err := v.Unmarshal(&cfg)
	if err != nil {
		logger.Info(err)
		return nil
	}

	// unmarshal component
	err = v.Unmarshal(&cfg.CommonConfig)
	if err != nil {
		logger.Info(err)
		return nil
	}

	//
	bannedAuthors := mapset.NewSet()
	if cfg.BannedAuthors != "" {
		for _, str := range strings.Split(cfg.BannedAuthors, "|") {
			bannedAuthors.Add(str)
		}
	}

	bannedTitles := mapset.NewSet()
	if cfg.BannedTitles != "" {
		for _, str := range strings.Split(cfg.BannedTitles, "|") {
			bannedTitles.Add(str)
		}
	}

	logger.Info(cfg)
	return &SmthRentInsighter{
		Config:        cfg,
		authorSet:     mapset.NewSet(),
		bannedAuthors: bannedAuthors,
		bannedTitles:  bannedTitles,
	}
}

// SmthRentInsighter ...
type SmthRentInsighter struct {
	Config config.SmthRentConfig

	pageURLs  []string
	authorSet mapset.Set
	client    fasthttp.Client

	bannedAuthors mapset.Set
	bannedTitles  mapset.Set
}

// Insight - insight smth rent
func (s *SmthRentInsighter) Insight(ctx context.Context) {
	defer logger.Sync()

	//
	var dataList []*SmthData

	// Instantiate default collector
	c := colly.NewCollector()

	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	c.CacheDir = s.Config.CacheDir

	if s.Config.NewCache {
		os.RemoveAll(c.CacheDir)
	}

	//
	err := s.getPageURLs()
	if err != nil || len(s.pageURLs) == 0 {
		logger.Infow("fail to get page list", "pageURLs", len(s.pageURLs), "error", err)
		return
	}

	var u, _ = url.Parse(s.Config.URL)
	var domainURL = u.Scheme + "://" + u.Host

	// OnHTML must be set before Visit
	// Parse html to get info
	c.OnHTML("#main #body .b-content table tbody tr:not(.ad)", func(e *colly.HTMLElement) {
		txn := config.DB.NewTransaction(true)
		defer txn.Discard()

		data := &SmthData{}
		data.populate(e.DOM, domainURL)

		if !s.isValid(data) {
			return
		}

		// Check whether or not in database
		key := data.Title + "_" + data.Author
		item, err := txn.Get([]byte(key))
		if err == nil {
			_, e := item.Value()
			if e == nil {
				return
			}
		}

		// add data to datalist
		dataList = append(dataList, data)
		err = txn.Set([]byte(key), []byte("0"), byte(0))
		if err != nil {
			logger.Info(err.Error())
		}
	})

	// Start scrapping
	for _, url := range s.pageURLs {
		go func(u string) {
			c.Visit(u)
		}(url)
	}
	c.Wait()

	// Output result
	if len(dataList) == 0 {
		logger.Info("final rent data is empty")
		return
	}

	if s.Config.NewDownload {
		os.RemoveAll(s.Config.DownloadDir)
	}

	filename := filepath.Join(s.Config.DownloadDir, "smth_"+time.Now().Format("20060102150405"))
	logger.Infow("fetching completed", "total_number", len(dataList), "filename", filename)
	// err = s.outputCSV(filename, dataList)
	s.outputXLSX(filename, dataList)

	if err != nil {
		logger.Infow("output result error", "error", err)
	}
}

func (s *SmthRentInsighter) getPageURLs() error {
	homePage := s.Config.URL + "&p=1"
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

	numStr := doc.Find("#body div.t-pre ul.pagination ol.page-main li:nth-last-child(2) a").Text()
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return err
	}

	if num > 0 {
		for k := 0; k < num; k++ {
			requestURL := s.Config.URL + "&p=" + strconv.Itoa(k+1)
			s.pageURLs = append(s.pageURLs, requestURL)
		}
	}

	return nil
}

func (s *SmthRentInsighter) isValid(data *SmthData) (v bool) {
	if data == nil {
		return false
	}

	v = true
	s.bannedTitles.Each(func(item interface{}) bool {
		if strings.Contains(data.Title, item.(string)) {
			v = false
			return true
		}
		return false
	})

	if !v {
		return
	}

	if s.bannedAuthors.Contains(data.Author) || s.authorSet.Contains(data.Author) {
		v = false
		return
	}

	s.authorSet.Add(data.Author)
	if data.Last.AddDate(0, 0, 15).Before(time.Now()) {
		v = false
		return
	}

	return
}

func (s *SmthRentInsighter) outputCSV(filename string, dataList []*SmthData) error {
	fp := filename + ".csv"
	file, err := util.CreateFile(fp)
	if err != nil {
		return err
	}
	return gocsv.MarshalFile(&dataList, file)
}

func (s *SmthRentInsighter) outputXLSX(filename string, dataList []*SmthData) error {
	var err error
	fp := filename + ".xlsx"

	baseDir := filepath.Dir(fp)
	if exists, err := util.Exists(baseDir); !exists || (err != nil) {
		err = os.MkdirAll(baseDir, os.ModePerm)
		if err != nil {
			logger.Errorw("create base dir for saving result err", "error_msg", err, "base_dir", baseDir)
			return err
		}
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		logger.Infow("create sheet error", "info", err)
		return err
	}

	headerRow := sheet.AddRow()
	headerRow.AddCell().SetValue("title")
	headerRow.AddCell().SetValue("href")
	headerRow.AddCell().SetValue("author")
	headerRow.AddCell().SetValue("comments")
	headerRow.AddCell().SetValue("last")

	for _, data := range dataList {
		row := sheet.AddRow()

		//
		row.AddCell().SetValue(data.Title)
		row.AddCell().SetValue(data.Href)
		row.AddCell().SetValue(data.Author)
		row.AddCell().SetValue(data.Comments)
		row.AddCell().SetDateTime(data.Last)
	}

	err = file.Save(fp)

	if err != nil {
		logger.Infow("save result to xlsx error", "error_msg", err)
	}

	return err
}
