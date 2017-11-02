package main

import (
	"time"

	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/router"
	"github.com/shohi/goinsight/util"
	"golang.org/x/net/context"
)

func main() {
	defer util.LogProcessTime(time.Now())

	ctx := context.WithCacel(context.Background())

	// load config
	config.Init(ctx)

	// route
	router.Route(ctx)
}
