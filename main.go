package main

import (
	"flag"
	"strings"

	"github.com/shohi/goinsight/basic"
	"github.com/shohi/goinsight/config"
)

func main() {

	flag.Parse()

	if strings.Compare(config.MainURL, "") == 0 {
		panic("url should be set!")
	}

	basic.InsightJSONImage(config.MainURL)
}
