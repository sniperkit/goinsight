package basic

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFetchTags(t *testing.T) {
	err := DoubanInsighter.fetchTags()

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("len: %d, tags: %v\n", len(DoubanInsighter.Tags), DoubanInsighter.Tags)
	}
}

func TestBytes(t *testing.T) {
	aa := []byte{0x01, 0x02}

	bb := make([]byte, 256)

	fmt.Println(bytes.NewBuffer(aa).Read(bb))
}

func TestGetSubjectID(t *testing.T) {
	url := "https://book.douban.com/subject/6082808/"

	got := DoubanInsighter.getSubjectID(url)
	want := "6082808"
	if got != want {
		t.Errorf("getSubjectID(%q) == %v, want %v", url, got, want)
	}
}
