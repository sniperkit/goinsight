package rent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/asciimoo/colly"
	"github.com/deckarep/golang-set"
	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/util"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
	"github.com/valyala/fasthttp"
)

// TcData ...
type TcData struct {
	Title    string    `csv:"title"`
	Rental   float64   `csv:"rental"`
	Room     string    `csv:"Room"`
	District string    `csv:"district"`
	Address  string    `csv:"address"`
	Href     string    `csv:"href"`
	Landlord string    `csv:"landlord"`
	Last     time.Time `csv:"last"`
}

func (d *TcData) populate(s *goquery.Selection) (err error) {
	// time
	sortid, exists := s.Attr("sortid")
	if exists {
		ut, err := strconv.ParseInt(sortid, 10, 64)
		if err != nil {
			return err
		}

		d.Last = time.Unix(ut/1000, ut%1000)
	} else {
		err = errors.New("Time Parse Error")
		return
	}

	// title
	d.Title = strings.Trim(s.Find(".des a").First().Text(), " ")

	href, exists := s.Find(".des a").First().Attr("href")
	if exists {
		d.Href = strings.Trim(href, " ")
	} else {
		err = errors.New("Link Not Found")
		return
	}

	d.Room = strings.Trim(s.Find("p.room").Text(), " ")
	d.Landlord = strings.Trim(s.Find(".des .geren").Text(), " ")
	d.Address = strings.Trim(s.Find(".des p.add").Text(), " ")
	d.District = strings.Fields(d.Address)[0]

	d.Rental, err = strconv.ParseFloat(s.Find(".listliright .money b").Text(), 64)
	return
}

// NewTcRentInsighter -- create new TcRentInsighter using configuration
func NewTcRentInsighter(v *viper.Viper) *TcRentInsighter {
	var cfg config.TcRentConfig

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
	allowedDistricts := mapset.NewSet()
	if cfg.AllowedDistricts != "" {
		for _, str := range strings.Split(cfg.AllowedDistricts, "|") {
			allowedDistricts.Add(str)
		}
	}

	bannedRooms := mapset.NewSet()
	if cfg.BannedRooms != "" {
		for _, str := range strings.Split(cfg.BannedRooms, "|") {
			bannedRooms.Add(str)
		}
	}

	logger.Info(cfg)
	return &TcRentInsighter{
		Config:           cfg,
		allowedDistricts: allowedDistricts,
		bannedRooms:      bannedRooms,
	}
}

// TcRentInsighter ...
type TcRentInsighter struct {
	Config config.TcRentConfig

	pageURLs []string
	client   fasthttp.Client

	allowedDistricts mapset.Set
	bannedRooms      mapset.Set
	stopFlag         bool // if response url is different from request's, we need to stop fetching
}

func (s *TcRentInsighter) getPageURLs() error {
	homePage := fmt.Sprintf(s.Config.URL, 1)

	statusCode, body, err := s.client.Get(nil, homePage)

	logger.Info("home page", homePage)

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

	numStr := doc.Find("#bottom_ad_li a:not(.next, .prv) span").Last().Text()
	num, err := strconv.Atoi(numStr)

	if err != nil {
		num = s.Config.DefaultTotalPages
	}

	for k := 0; k < num; k++ {
		requestURL := fmt.Sprintf(s.Config.URL, k+1)
		s.pageURLs = append(s.pageURLs, requestURL)
	}

	return nil
}

// Insight - insight 58tongcheng rent
// implement interface
func (s *TcRentInsighter) Insight(ctx context.Context) {
	defer logger.Sync()

	//
	var dataList []*TcData

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

	// OnHTML must be set before Visit
	// Parse html to get info
	c.OnHTML(".main .content .listBox .listUl>li[logr][sortid]", func(e *colly.HTMLElement) {
		txn := config.DB.NewTransaction(true)
		defer txn.Discard()

		if s.stopFlag {
			return
		}

		data := &TcData{}
		data.populate(e.DOM)

		if !s.isValid(data) {
			return
		}

		// Check whether or not in database
		key := data.Title + "_" + data.Room
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

	c.OnRequest(func(req *colly.Request) {
		req.Ctx.Put("OriginURL", req.URL.String())
	})

	c.OnResponse(func(res *colly.Response) {
		originURL := res.Ctx.Get("OriginURL")
		if strings.Compare(res.Request.URL.String(), originURL) != 0 {
			s.stopFlag = true
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

	filename := filepath.Join(s.Config.DownloadDir, "tc_"+time.Now().Format("20060102150405"))
	s.outputXLSX(filename, dataList)

	if err != nil {
		logger.Infow("output result error", "error", err)
	}
}

func (s *TcRentInsighter) isValid(d *TcData) (v bool) {
	if d == nil {
		return false
	}

	v = true
	if d.Last.AddDate(0, 0, 15).Before(time.Now()) {
		v = false
		return
	}

	if s.bannedRooms.Contains(d.Room) {
		v = false
		return
	}

	if s.allowedDistricts.Contains(d.District) {
		v = true
		return
	}

	return
}

func (s *TcRentInsighter) outputXLSX(filename string, dataList []*TcData) error {
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

	headRow := sheet.AddRow()
	headRow.AddCell().SetValue("title")
	headRow.AddCell().SetValue("rental")
	headRow.AddCell().SetValue("room")
	headRow.AddCell().SetValue("district")
	headRow.AddCell().SetValue("address")
	headRow.AddCell().SetValue("href")
	headRow.AddCell().SetValue("landlord")
	headRow.AddCell().SetValue("last")

	for _, data := range dataList {
		row := sheet.AddRow()
		//
		row.AddCell().SetValue(data.Title)
		row.AddCell().SetValue(data.Rental)
		row.AddCell().SetValue(data.Room)
		row.AddCell().SetValue(data.District)
		row.AddCell().SetValue(data.Address)
		row.AddCell().SetValue(data.Href)
		row.AddCell().SetValue(data.Landlord)
		row.AddCell().SetDateTime(data.Last)
	}

	err = file.Save(fp)

	return err
}
