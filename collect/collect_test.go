package collect

import (
	"testing"
	"time"

	"github.com/dreamerjackson/crawler/proxy"
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

func Test_BrowserFetchWithTimeout(t *testing.T) {
	url := "https://book.douban.com/subject/1007305/"
	req := &spider.Request{
		URL: url,
	}
	f := BrowserFetch{
		Timeout: 300 * time.Millisecond,
	}
	body, err := f.Get(req)
	if err != nil {
		// log err info : timeout exceed
		t.Log(err.Error())
	}
	t.Log(string(body))
}

func Test_BrowserFetchWithProxy(t *testing.T) {
	// url that you host can't access directly
	url := "https://www.google.com/"
	// your proxy server address
	proxyURLs := []string{"http://127.0.0.1:7890"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		t.Fatal("RoundRobinProxySwitcher failed")
	}
	req := &spider.Request{
		URL: url,
	}
	f := BrowserFetch{
		Proxy:   p,
		Timeout: 40 * time.Second,
	}
	body, err := f.Get(req)
	assert.Nil(t, err)
	t.Log(string(body))
}
