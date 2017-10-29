// Package config - contains all global configurations that should be set once at the startup
package config

import "flag"

var (
	// MainURL - the first entry url when scrapping
	MainURL string

	// DirName - the directory to save downloaded files
	DirName string
)

func init() {

	flag.StringVar(&MainURL, "url", "", "entry url for scrapping")
	flag.StringVar(&DirName, "dir", "_dl", "download directory")
}
