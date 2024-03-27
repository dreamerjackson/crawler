package economist

import (
	"github.com/dreamerjackson/crawler/proxy"
	"github.com/dreamerjackson/crawler/spider"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseEconomistList(t *testing.T) {
	p, err := proxy.RoundRobinProxySwitcher("http://127.0.0.1:8888")
	assert.Nil(t, err)

	f := spider.NewFetchService(spider.BrowserFetchType)
	//task.Logger = c.Logger
	rootreqs, err := EconomistTask.Rule.Root()
	EconomistTask.Proxy = p
	assert.Nil(t, err)

	var reqs []*spider.Request
	reqs = append(reqs, rootreqs...)

	for len(reqs) > 0 {

		req := reqs[0]
		reqs = reqs[1:]
		req.Task = EconomistTask
		body, err := f.Get(req)
		assert.NoError(t, err)
		if len(body) < 6000 {
			t.Logf("can't fetch length:%v url:%v", len(body), req.URL)
			continue
		}
		t.Log("start visit: ", req.URL, "body length:", len(body))

		time.Sleep(1 * time.Second)
		rule := req.Task.Rule.Trunk[req.RuleName]
		ctx := &spider.Context{
			Body: body,
			Req:  req,
		}
		result, err := rule.ParseFunc(ctx)
		assert.Nil(t, err)
		if len(result.Items) > 0 {
			t.Log("result:", result.Items)
		}
		if len(result.Requesrts) > 0 {
			t.Logf("add result %+v", result.Requesrts[0])
			reqs = append(reqs, result.Requesrts...)
		}
	}

}
