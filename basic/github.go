package basic

import (
	"fmt"
	"strings"

	"github.com/asciimoo/colly"
)

// InsightGithub - fetch github repos' stared and forking data
// and conclude some insights

var searchBaseURL = "https://github.com/search?q="
var sep, collapse = ":", "+"

func InsightGithub() {
	// Instantiate default collector
	c := colly.NewCollector()

	// Visit only domains: douban.com, book.douban.com
	c.AllowedDomains = []string{"github.com"}

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

	// Start scraping on github.com
	m := map[string]string{
		"language": "go",
		"stars":    ">500",
	}
	c.Visit(searchBaseURL + joinMap(m))

}

func joinMap(m map[string]string) string {
	var entries []string
	for k, v := range m {
		entries = append(entries, k+sep+v)
	}

	return strings.Join(entries, collapse)
}
