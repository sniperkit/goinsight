// Package router -- route insight request based on type
package router

import (
	"context"

	"github.com/shohi/goinsight/basic"
	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/special/rent"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var logger = zap.NewExample().Sugar()

// Route - route url from configuration
func Route(ctx context.Context) {
	var insighter basic.Insighter
	t := config.BaseConfig.Type

	if t == "rent-smth" {
		insighter = rent.NewSmthRentInsighter(viper.Sub(t))
	} else if t == "rent-tc" {
		insighter = rent.NewTcRentInsighter(viper.Sub(t))
	} else {
		insighter = basic.NewJSONImageInsighter(viper.Sub(t))
	}

	logger.Infow("route", "type", t)
	insighter.Insight(ctx)
}
