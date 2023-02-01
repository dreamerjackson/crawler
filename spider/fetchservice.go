package spider

import (
	"bufio"
	"context"
	"fmt"
	"github.com/dreamerjackson/crawler/extensions"
	"go.uber.org/zap"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type FetchType int

const (
	BaseFetchType FetchType = iota
	BrowserFetchType
)

type Fetcher interface {
	Get(url *Request) ([]byte, error)
}

func NewFetchService(typ FetchType) Fetcher {
	switch typ {
	case BaseFetchType:
		return &baseFetch{}
	case BrowserFetchType:
		return &browserFetch{}
	default:
		return &browserFetch{}
	}
}

type baseFetch struct{}

func (*baseFetch) Get(req *Request) ([]byte, error) {
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

type browserFetch struct{}

// 模拟浏览器访问
func (b *browserFetch) Get(request *Request) ([]byte, error) {
	task := request.Task
	if err := task.Limit.Wait(context.Background()); err != nil {
		return nil, err
	}
	// 随机休眠，模拟人类行为
	sleeptime := rand.Int63n(task.WaitTime * 1000)
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)

	client := &http.Client{
		Timeout: task.Timeout,
	}

	if task.Proxy != nil {
		transport := http.DefaultTransport.(*http.Transport)
		transport.Proxy = task.Proxy
		client.Transport = transport
	}

	req, err := http.NewRequest("GET", request.URL, nil)

	if err != nil {
		return nil, fmt.Errorf("get url failed:%w", err)
	}

	if len(task.Cookie) > 0 {
		req.Header.Set("Cookie", task.Cookie)
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
