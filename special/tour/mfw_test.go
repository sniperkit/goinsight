package tour

import (
	"log"
	"testing"

	"github.com/shohi/goinsight/config"
)

func TestGetPages(t *testing.T) {
	cfg := config.CommonConfig{}
	cfg.URL = "http://www.mafengwo.cn/yj/10176/1-0-%d.html"

	s := &MfwTourInsighter{Config: cfg}
	log.Println(s.getPageURLs())
	log.Println(len(s.pageURLs))
}
