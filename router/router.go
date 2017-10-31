// Package router -- route insight request based on type
package router

import "github.com/shohi/goinsight/basic"

func Route(t, url string) {
	var insighter basic.Insighter
	switch t {
	case "image":
		insighter = &basic.ImageInsighter{}
	case "image-json":
		insighter = &basic.ImageInsighter{}
	case "github":
		insighter = basic.GithubInsighter
	case "book":
		insighter = &basic.BookInsighter{}
	default:
		insighter = nil
	}

	insighter.Insight(url)
}
