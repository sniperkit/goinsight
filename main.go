package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/router"
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

	router.Route(config.Type, config.MainURL)
}
