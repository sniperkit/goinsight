// Package router -- route insight request based on type
package router

import (
	"golang.org/x/net/context"

	"github.com/shohi/goinsight/basic"
	"github.com/shohi/goinsight/config"
	"github.com/spf13/viper"
)

func Route(ctx context.Context) {
	var insighter basic.Insighter
	t := config.BaseConfig.Type
	insighter = &basic.ImageInsighter{Config: viper.Sub(t)}
	insighter.Insight()
}
