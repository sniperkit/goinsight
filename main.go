package main

import (
	"context"
	"time"

	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/router"
	"github.com/shohi/goinsight/util"
	"go.uber.org/zap"
)

var logger = zap.NewExample()

func main() {
	defer util.LogProcessTime(logger, time.Now())

	ctx, doCancelFunc := context.WithCancel(context.Background())
	defer doCancelFunc()

	// load config
	config.Init(ctx)

	// route
	router.Route(ctx)
}
