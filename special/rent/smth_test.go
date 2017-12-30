package rent

import (
	"log"
	"testing"

	"github.com/shohi/goinsight/config"
)

func testGetPageURLs(t *testing.T) {

	cfg := config.SmthRentConfig{}
	cfg.URL = "http://www.newsmth.net/nForum/board/HouseRent?ajax"

	var s = &SmthRentInsighter{Config: cfg}
	log.Println(s.getPageURLs())
	log.Println(len(s.pageURLs))
	log.Println(s.pageURLs[0])
}
