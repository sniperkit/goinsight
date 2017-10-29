package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/shohi/goinsight/basic"
	"github.com/shohi/goinsight/config"
)

func main() {

	flag.Parse()
	startT := time.Now()
	defer func() {
		endT := time.Now()
		fmt.Println("process using ", endT.Sub(startT))
	}()

	if strings.Compare(config.MainURL, "") == 0 {
		panic("url should be set!")
	}

	basic.InsightJSONImage(config.MainURL)
}
