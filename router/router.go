// Package router -- route insight request based on type
package router

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
	"github.com/shohi/goinsight/basic"
)

func Route(t, url string) {

	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions
	opts.Dir = "/tmp/badger"
	opts.ValueDir = "/tmp/badger"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var insighter basic.Insighter
	switch t {
	case "image":
		insighter = &basic.ImageInsighter{}
	case "image-json":
		insighter = &basic.JSONImageInsighter{}
	case "github":
		insighter = basic.GithubInsighter
	case "book":
		insighter = basic.DoubanInsighter
	default:
		insighter = nil
	}

	fmt.Println(insighter)
	insighter.Insight(url)
}
