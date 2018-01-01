package rent

import (
	"log"
	"testing"

	"github.com/shohi/goinsight/config"
)

func testGetPages(t *testing.T) {
	cfg := config.TcRentConfig{}
	cfg.URL = "http://bj.58.com/chaoyang/hezu/0/pn%d/?minprice=1800_4000"

	var s = &TcRentInsighter{Config: cfg}
	log.Println(s.getPageURLs())
	log.Println(len(s.pageURLs))
	log.Println(s.pageURLs[0])
}
