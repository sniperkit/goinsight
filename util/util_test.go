package util

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/shohi/goinsight/model"
)

func TestStringTrim(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"a/b", "a/b"},
		{"/a/b/", "a/b/"},
	}

	for _, c := range cases {
		got := strings.TrimPrefix(c.input, "/")
		if strings.Compare(got, c.want) != 0 {
			t.Errorf("TrimPrefix(%q) == %v, want %v", c.input, got, c.want)
		}
	}
}

func TestStringSplit(t *testing.T) {
	fmt.Println(strings.Split("a/b/c.png", "/"))
}

func TestWriteOnly(t *testing.T) {
	f, err := CreateFile("tmp/test.tmp")
	defer f.Close()

	f.Write([]byte("hello"))

	if err != nil {
		t.Error(err)
	}
}

func TestDownload(t *testing.T) {
	url := "https://previews.123rf.com/images/benjaminboeckle/benjaminboeckle1611/benjaminboeckle161100512/67028130-Cape-of-good-Hope-in-South-Africa-Stock-Photo.jpg"
	err := Download(url, "tmp")
	fmt.Println(err)
}

func TestJSONMarshal(t *testing.T) {
	c := model.ImageCollection{
		Status: 1,
	}

	data, _ := json.Marshal(c)
	fmt.Println(string(data))
}

func TestURLHost(t *testing.T) {
	u, _ := url.Parse("http://localhost:9090/app")
	fmt.Println(u.Scheme)
}

func TestURLParse(t *testing.T) {
	baseURL := "http://localhost:9090/app/list?page=1&action=list"

	u, _ := url.Parse(baseURL)

	fmt.Println(u.Scheme)
	fmt.Println(u.Host)
	fmt.Println(u.Path)
	fmt.Println(u.Fragment)
	fmt.Println(u.RawQuery)

	m, _ := url.ParseQuery(u.RawQuery)
	fmt.Println(m["page"][0])
}

func TestGetPageURLs(t *testing.T) {
	baseURL := "http://localhost:9090/app/list?page=1&action=list"
	urls := GetPageURLs(baseURL, 10)
	fmt.Println(urls)
}

func TestGetPageParamter(t *testing.T) {
	baseURL := "http://localhost:9090/app/list?action=list"
	page := GetPageParamter(baseURL)

	fmt.Println(page)
}
