// Package util - contains common utils function
package util

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var client = &fasthttp.Client{}

// Download download files from given url
func Download(url, fp string, overwrite bool) error {
	if exists, _ := Exists(fp); exists && !overwrite {
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

// Exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
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

// LogProcessTime - log process time
func LogProcessTime(logger *zap.Logger, startT time.Time) {
	endT := time.Now()
	logger.Info("", zap.String("process time", endT.Sub(startT).String()))
}

// GetContent - get content directed by url
func GetContent(url string) ([]byte, error) {
	statusCode, body, err := client.Get(nil, url)

	if err != nil {
		return nil, err
	}

	if statusCode != fasthttp.StatusOK {
		return nil, errors.New("Status Code Is Not OK")
	}

	return body, nil
}

// GetResourceName - get resource name in uri
func GetResourceName(uri string) (string, error) {
	rawURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	strs := strings.Split(rawURL.Path, "/")
	lastItem := strs[len(strs)-1]

	return strings.Split(lastItem, ".")[0], nil
}

// GetResourceSuffix - get resource suffix in uri, including `.`
func GetResourceSuffix(uri string) (string, error) {

	rawURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	strs := strings.Split(rawURL.Path, "/")
	lastItem := strs[len(strs)-1]
	subItems := strings.Split(lastItem, ".")

	return "." + subItems[len(subItems)-1], nil
}
