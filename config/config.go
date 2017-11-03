// Package config - contains all global configurations that should be set once at the startup
package config

import (
	"os"

	"golang.org/x/net/context"

	"github.com/dgraph-io/badger"
	"github.com/shohi/goinsight/util"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type baseConfig struct {
	// search type
	Type string
}

type badgerConfig struct {
	Dir      string
	ValueDir string
}

type CommonConfig struct {
	// entry url for scrapping
	URL string

	// download directory
	DownloadDir string

	CacheDir string
}

type BookConfig struct {
	CommonConfig
}

type JSONImageConfig struct {
	URL         string
	DownloadDir string
	CacheDir    string
	ThresHold   int
}

type ImageConfig struct {
	CommonConfig
}

var (
	BaseConfig   baseConfig
	BadgerConfig badgerConfig
	DB           *badger.DB
)

var logger = zap.NewExample().Sugar()

// Init - load configs from file
func Init(ctx context.Context) {
	err := loadTOML()
	if err != nil {
		panic(err)
	}

	//
	baseCfg := viper.Sub("base")
	baseCfg.Unmarshal(&BaseConfig)

	badgerConfig := viper.Sub("badger")
	badgerConfig.Unmarshal(&BadgerConfig)

	//
	initDB(ctx)

}

func loadTOML() error {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// InitDB - init badger
func initDB(ctx context.Context) {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.

	opts := badger.DefaultOptions
	opts.Dir = BadgerConfig.Dir
	opts.ValueDir = BadgerConfig.ValueDir

	if exists, err := util.Exists(opts.Dir); !exists || (err != nil) {
		err = os.MkdirAll(opts.Dir, os.ModePerm)
		if err != nil {
			logger.Fatal(err)
			panic("Init DB Error")
		}
	}

	db, err := badger.Open(opts)
	if err != nil {
		logger.Fatal(err)
	}

	// close db when context is done
	go func() {
		select {
		case <-ctx.Done():
			logger.Info("DB to be closed")
			db.Close()
		}
	}()

	DB = db
}
