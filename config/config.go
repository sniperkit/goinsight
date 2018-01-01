// Package config - contains all global configurations that should be set once at the startup
package config

import (
	"context"
	"os"

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

// CommonConfig - common config
type CommonConfig struct {
	// entry url for scrapping
	URL string

	// download directory
	DownloadDir string

	CacheDir string
}

// BookConfig - configuration for book info scrapping
type BookConfig struct {
	CommonConfig
}

// JSONImageConfig - configuration for downloading image whose info is in json format
type JSONImageConfig struct {
	CommonConfig
	ThresHold int
}

// MfwImageConfig - configuration for downloading info from mfw
type MfwImageConfig struct {
	CommonConfig
}

// ImageConfig - configuration for downloading image from regular pages
type ImageConfig struct {
	CommonConfig
}

// SmthRentConfig - configuration for fetching rent information from SMTH
type SmthRentConfig struct {
	CommonConfig
	BannedAuthors string
	BannedTitles  string
}

// TcRentConfig - configuration for fetching rent information from `58同城`
type TcRentConfig struct {
	CommonConfig

	AllowedDistricts  string
	BannedRooms       string
	DefaultTotalPages int
}

var (
	// BaseConfig - config type
	BaseConfig baseConfig

	// BadgerConfig - config of badger
	BadgerConfig badgerConfig

	// DB - badger DB
	DB *badger.DB
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
