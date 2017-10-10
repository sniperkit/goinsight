package basic

import (
	"fmt"

	"github.com/asciimoo/colly"
)

// InsightBook fetch book data from douban and conclude some insights
func InsightBook() {

	// Instantiate default collector
	c := colly.NewCollector()

	// Visit only domains: douban.com, book.douban.com
	c.AllowedDomains = []string{"douban.com", "book.douban.com"}

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

	// Start scraping on https://hackerspaces.org
	c.Visit("https://book.douban.com")
}
