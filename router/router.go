// Package router -- route insight request based on type
package router

import (
	"context"

	"github.com/shohi/goinsight/basic"
	"github.com/shohi/goinsight/config"
	"github.com/spf13/viper"
)

// Route - route url from configuration
func Route(ctx context.Context) {
	var insighter basic.Insighter
	t := config.BaseConfig.Type
	insighter = basic.NewJSONImageInsighter(viper.Sub(t))
	insighter.Insight(ctx)
}
