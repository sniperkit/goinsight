// Package router -- route insight request based on type
package router

import (
	"fmt"

	"github.com/shohi/goinsight/basic"
)

func Route(t, url string) {
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
