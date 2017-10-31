package basic

import (
	"fmt"
	"strings"

	"github.com/asciimoo/colly"
)

type cvsInsighter struct {
	BaseURL  string
	Sep      string
	Collapse string
}

// GithubInsighter -- search repos in github
var GithubInsighter = &cvsInsighter{"https://github.com/search?q=", ":", "+"}

// InsightGithub - fetch github repos' stared and forking data
// and conclude some insights
func (i *cvsInsighter) Insight(_ string) {
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
	c.Visit(i.BaseURL + i.joinMap(m))

}

func (i *cvsInsighter) joinMap(m map[string]string) string {
	var entries []string
	for k, v := range m {
		entries = append(entries, k+i.Sep+v)
	}

	return strings.Join(entries, i.Collapse)
}
