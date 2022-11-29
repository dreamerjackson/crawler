package collect

import (
	"bufio"
	"fmt"
	"github.com/dreamerjackson/crawler/spider"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dreamerjackson/crawler/extensions"
	"github.com/dreamerjackson/crawler/proxy"
	"go.uber.org/zap"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type BaseFetch struct{}

func (BaseFetch) Get(req *spider.Request) ([]byte, error) {
	resp, err := http.Get(req.URL)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status code:%d", resp.StatusCode)
	}

	bodyReader := bufio.NewReader(resp.Body)
	e := DeterminEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}

type BrowserFetch struct {
	Timeout time.Duration
	Proxy   proxy.Func
	Logger  *zap.Logger
}

// 模拟浏览器访问
func (b BrowserFetch) Get(request *spider.Request) ([]byte, error) {
	client := &http.Client{
		Timeout: b.Timeout,
	}

	if b.Proxy != nil {
		transport := http.DefaultTransport.(*http.Transport)
		transport.Proxy = b.Proxy
		client.Transport = transport
	}

	req, err := http.NewRequest("GET", request.URL, nil)

	if err != nil {
		return nil, fmt.Errorf("get url failed:%w", err)
	}

	if len(request.Task.Cookie) > 0 {
		req.Header.Set("Cookie", request.Task.Cookie)
	}

	req.Header.Set("User-Agent", extensions.GenerateRandomUA())

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	bodyReader := bufio.NewReader(resp.Body)
	e := DeterminEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}

func DeterminEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)

	if err != nil {
		zap.L().Error("fetch failed", zap.Error(err))

		return unicode.UTF8
	}

	e, _, _ := charset.DetermineEncoding(bytes, "")

	return e
}
