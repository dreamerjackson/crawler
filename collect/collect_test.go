package collect

import (
	"testing"

	"github.com/dreamerjackson/crawler/spider"
	"github.com/stretchr/testify/assert"
)

func Test_BaseFetch(t *testing.T) {

	url := "https://book.douban.com/subject/1007305/"
	req := &spider.Request{
		URL: url,
	}
	f := BaseFetch{}
	body, err := f.Get(req)
	assert.Nil(t, body)
	t.Log(err.Error())
}

func Test_BrowserFetch(t *testing.T) {
	url := "https://book.douban.com/subject/1007305/"
	req := &spider.Request{
		URL: url,
	}
	f := BrowserFetch{}
	body, err := f.Get(req)
	assert.Nil(t, err)
	t.Log(string(body))
}
