package config

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoadTOML(t *testing.T) {
	ctx, doCancelFunc := context.WithCancel(context.Background())
	defer func() {
		doCancelFunc()
		time.Sleep(time.Second)
	}()
	Init(ctx)
	log.Println(BaseConfig)
	log.Println(BadgerConfig)

	// unmarshal direct fields
	var cfg SmthRentConfig
	vs := viper.Sub("rent-smth")
	vs.Unmarshal(&cfg)

	// unmarshal component
	err := vs.Unmarshal(&cfg.CommonConfig)

	if err != nil {
		log.Println(err)
	}
	log.Println(cfg)
}
