package basic

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/asciimoo/colly"
)

// reference, https://github.com/hunterhug/doubanbook30000
// book:
// http://book.douban.com/tag/
// http://www.douban.com/tag/小说/book?start=0 书列表 间隔15
// http://book.douban.com/subject/25862578/?from=tag_all 书信息
// http://book.douban.com/subject/6082808/reviews?score=&start=0 书评 间隔25
// https://book.douban.com/subject/6082808/doulists 豆列

// author:
// https://book.douban.com/author/1039386/
// https://book.douban.com/author/1039386/books?sortby=time&format=pic

type bookInsighter struct {
	URL     string
	Tags    []string
	Domains []string
}

// Book - book info
type Book struct {
	URL       string
	SubjectID string
	Title     string
	ImageURL  string

	Author      string
	OriginTitle string
	Publisher   string
	Translator  string
	PubYear     string
	Pages       int
	Price       float32
	Binding     string
	Series      string
	ISBN        string

	Rate      int // 100-based
	RateUsers int // 1000-based
	FiveStar  int
	FourStar  int
	ThreeStat int
	TwoStar   int
	OneStar   int

	Tags         []string
	Similarities []string

	Doulists      int
	ShortComments int
	LongComments  int
	Notes         int

	SecondHands int
	IsReading   int
	HasReaded   int
	WantReading int

	JdPrice  float32
	DdPrice  float32
	AmzPrice float32

	Libraries []string
	Versions  int
}

// DoubanInsighter - fetch book data from douban and conclude some insights
var DoubanInsighter = &bookInsighter{URL: "https://book.douban.com/tag/"}

// Insight - fetch book data
func (i *bookInsighter) Insight(_ string) {

	// Instantiate default collector
	c := colly.NewCollector()

	// Visit only domains: douban.com, book.douban.com
	if len(i.Domains) > 0 {
		c.AllowedDomains = i.Domains
	}

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// c.OnResponse(

	// Start scraping on https://book.douban.com
	i.fetchTags()
	for _, v := range i.Tags {
		go func(l string) {
			c.Visit(i.URL + l)
		}(v)
	}

	c.Wait()
}

// ref https://github.com/PuerkitoBio/goquery for goquery's details
func (i *bookInsighter) fetchTags() error {
	// Download html file
	doc, err := goquery.NewDocument(i.URL)
	if err != nil {
		return err
	}

	// Find book tags in html
	var tags []string
	doc.Find("table.tagCol tbody tr td a").Each(func(i int, s *goquery.Selection) {
		tag := s.Text()
		if tag != "" {
			tags = append(tags, tag)
		}
	})

	i.Tags = tags

	return nil
}

func (i *bookInsighter) parseFromRespone(res *colly.Response) (*Book, error) {

	// Download html file
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(res.Body))
	if err != nil {
		return nil, err
	}

	//
	var book Book

	// Get Title and SubjectID
	book.URL = res.Request.URL.String()
	book.Title = doc.Find("#dale_book_subject_top_icon + h1 span").First().Text()
	book.SubjectID = i.getSubjectID(res.Request.URL.String())
	if src, exists := doc.Find("#mainpic a.nbg img").First().Attr("src"); exists {
		book.ImageURL = src
	}

	//

	return &book, nil
}

func (i *bookInsighter) getSubjectID(url string) string {
	var subjectID string
	ss := strings.Split(url, "/")
	for k := len(ss) - 1; k >= 0; k-- {
		if ss[k] != "" {
			subjectID = ss[k]
			break
		}
	}
	return subjectID
}
