// Package util - contains common utils function
package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shohi/goinsight/config"
	"github.com/shohi/goinsight/model"
	"github.com/valyala/fasthttp"
)

var client = &fasthttp.Client{}

// Download download files from given url
func Download(url string, subdir string) error {
	filename := FilenameFromURL(url)
	var fp string
	if strings.Compare(subdir, "") == 0 {
		fp = filepath.Join(config.DirName, filename)
	} else {
		fp = filepath.Join(config.DirName, subdir, filename)
	}

	if exists, _ := exists(fp); exists {
		log.Println(fp, " exists ")
		return nil
	}

	file, err := CreateFile(fp)
	if err != nil {
		return err
	}

	statusCode, body, err := client.Get(nil, url)
	if err != nil {
		return err
	}

	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("Status Code Is Not %d", fasthttp.StatusOK)
	}

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	return nil
}

// FilenameFromURL -- get file name from url, where url is in form of 'http://..../filename'
func FilenameFromURL(url string) string {
	ss := strings.Split(url, "/")
	return ss[len(ss)-1]
}

// CreateFile - create file anyway, if file exists, empty it and return
func CreateFile(fp string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(fp), os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0600)
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// LoadImageJSON -- loads image json info
func LoadImageJSON(url string) (interface{}, error) {

	client := &fasthttp.Client{}
	statusCode, body, err := client.Get(nil, url)

	if err != nil {
		return nil, err
	}

	if statusCode != fasthttp.StatusOK {
		return nil, errors.New("Status Code Is Not OK")
	}

	var info model.ImageCollection

	// The BOM identifies that the text is UTF-8 encoded, but it should be removed before decoding.
	// https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(body, &info)

	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetPageURLs - parse given url and generate several urls with different page number
// base url contains `page=index` param
func GetPageURLs(baseURL string, pages int, exclude bool) (urls []string) {
	if pages < 1 {
		return
	}

	u, _ := url.Parse(baseURL)
	m, _ := url.ParseQuery(u.RawQuery)
	basePage := GetPageParamter(baseURL)

	for k := 1; k <= pages; k++ {
		pageStr := strconv.Itoa(k)
		if exclude && strings.Compare(basePage, pageStr) == 0 {
			continue
		}
		m["page"] = []string{pageStr}
		urls = append(urls, u.Scheme+"://"+u.Host+u.Path+"?"+m.Encode())
	}

	return
}

// GetPageParamter get page number
func GetPageParamter(baseURL string) string {
	u, _ := url.Parse(baseURL)
	m, _ := url.ParseQuery(u.RawQuery)

	if len(m["page"]) < 1 {
		return ""
	}

	return m["page"][0]
}
