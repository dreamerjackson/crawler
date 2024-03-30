package bloomberg

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/dreamerjackson/crawler/proxy"
	"github.com/dreamerjackson/crawler/spider"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestChomeDp2(t *testing.T) {
	// 创建一个chromedp上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 启用网络
	err := chromedp.Run(ctx, network.Enable())
	if err != nil {
		fmt.Println(err)
	}

	var cookies []*network.Cookie
	// 导航到页面并获取cookies
	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.bloomberg.com/ai`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Cookies:")
	for _, cookie := range cookies {
		fmt.Printf("- %s: %s\n", cookie.Name, cookie.Value)
	}
}

// splitCookies 用于从一个由分号分割的字符串中提取并返回多个 CookieParam
func splitCookies(cookieStr string) []*network.CookieParam {
	var cookies []*network.CookieParam
	for _, c := range strings.Split(cookieStr, "; ") {
		parts := strings.SplitN(c, "=", 2)
		if len(parts) == 2 {
			cookies = append(cookies, &network.CookieParam{Name: parts[0], Value: parts[1]})
		}
	}
	return cookies
}

var cookieStr string = `optimizelyEndUserId=oeu1711550399176r0.1877745177569461; bbgconsentstring=req1fun1pad1; bdfpc=004.9447305864.1711550400529; _gcl_au=1.1.1799522145.1711550401; consentUUID=c858b581-08ff-4283-bce4-67f4d9b469ba; _fbp=fb.1.1711550401204.1932059742; usnatUUID=f2ace540-bc32-45fa-9cd5-380fd71e788a; pxcts=e472f5df-ec47-11ee-b75b-1b6d6af0f63f; _pxvid=e472e655-ec47-11ee-b75b-4c8dfa9e3f3b; agent_id=1f931978-32eb-4c2b-adeb-56dabf0ee4ab; session_id=a15a91ba-c52e-4150-9967-ec9cc2b88264; session_key=9ad6d5dad31de9a68877b65a95b3d3ba0adca711; gatehouse_id=7c24689a-e9aa-4e70-8b65-5916faa74d28; geo_info=%7B%22countryCode%22%3A%22HK%22%2C%22country%22%3A%22HK%22%2C%22field_n%22%3A%22cp%22%2C%22trackingRegion%22%3A%22Asia%22%2C%22cacheExpiredTime%22%3A1712155201555%2C%22region%22%3A%22Asia%22%2C%22fieldN%22%3A%22cp%22%7D%7C1712155201555; exp_pref=APAC; country_code=HK; seen_uk=1; _gcl_aw=GCL.1711550412.CjwKCAjwh4-wBhB3EiwAeJsppNyuWpdTxz82Fs7ec8vFx5f8BAo40NNfkuhoEz1ePJJOF_yLTjtWLhoCGD4QAvD_BwE; _gcl_dc=GCL.1711550412.CjwKCAjwh4-wBhB3EiwAeJsppNyuWpdTxz82Fs7ec8vFx5f8BAo40NNfkuhoEz1ePJJOF_yLTjtWLhoCGD4QAvD_BwE; afUserId=f1d90c08-d5d5-4750-a181-58830064d649-p; AF_SYNC=1711550414351; _cc_id=4b00edd3be9c419a4ca180d3533f94f5; panoramaId=1fd98f3669d105fc90dc837acd15a9fb927ae6090142d1edb7c3b9439c610e4e; panoramaIdType=panoDevice; _scid=6385e024-98dc-4487-87b5-3e2da7c1b415; _sctr=1%7C1711468800000; __stripe_mid=7905422e-2c3f-4d4d-b502-cac5a0831e8be7dcb2; _scid_r=6385e024-98dc-4487-87b5-3e2da7c1b415; professional-cookieConsent=new-relic|perimeterx-bot-detection|perimeterx-pixel|google-tag-manager|google-analytics|microsoft-clarity|optimizely|microsoft-advertising|eloqua|adwords|linkedin-insights; _gid=GA1.2.267148422.1711556267; drift_aid=5f79dcc1-c67f-413f-bd72-f8397766c4cc; driftt_aid=5f79dcc1-c67f-413f-bd72-f8397766c4cc; _ga_NNP7N7T2TG=GS1.1.1711556267.1.1.1711556279.48.0.0; _ga=GA1.1.928448926.1711550401; __gads=ID=0adf99e208745bfb:T=1711550412:RT=1711600226:S=ALNI_MaftDuDZJA8HYObZ08pS8lwAPixcQ; __gpi=UID=00000d6ad874281c:T=1711550412:RT=1711600226:S=ALNI_MbaCozYcJsrtQwyEUxfIXAgIQN0CQ; __eoi=ID=39c1f3bd6f4a61b5:T=1711550412:RT=1711600226:S=AA-AfjZL1Az9HIYk5RZl4ilz7r9K; _clck=8d76t0%7C2%7Cfkg%7C0%7C1547; _parsely_session={%22sid%22:2%2C%22surl%22:%22https://www.bloomberg.com/asia%22%2C%22sref%22:%22%22%2C%22sts%22:1711600227224%2C%22slts%22:1711550401154}; _parsely_visitor={%22id%22:%22pid=7f1bee4ee7907a07f60ca606849702d6%22%2C%22session_count%22:2%2C%22last_session_ts%22:1711600227224}; _user-data=%7B%22status%22%3A%22anonymous%22%7D; __sppvid=36e9b221-932f-40cf-bc63-2f4007873f35; _uetsid=e419b980ec4711eea00277baae156ef6; _uetvid=fa9a1f20455411ec9b74dbfc533fddae; _reg-csrf-token=dWw8H8xh-YFIH7yxAWwE656GiBSM_5LJhDMc; _last-refresh=2024-3-28%204%3A30; _clsk=ro46fm%7C1711600234083%7C2%7C0%7Cn.clarity.ms%2Fcollect; _px3=3a49a8cecb63d99751073132d1f058c98a70f771eff02b1ce92cc76fcd979975:ToMyJqXIe4ITefCtC3LC7xJifq/LgdtbAKBF6Lj3fnt1u/hZMiyy+og+/r1V1CoE7q296ur52jJT0hFqvWwymw==:1000:YDyhkzul5GXL8980W1G1bH5GEVYWR58ma+myGto7jBcxh/ZOurLACbpahMPjXoD3YxYdO/KQ67rpzuoAy/5Tf/q0PVESof7jVCicrZyLhZSXXMq372zMRdMnPl4jqu4x+kLvfHPPw8kdN7FVcaD1Rb1fzWFDZaQKPGAi2BEpksWqrrVfa/AXFNpLpTRJhVYWxQ04hNBtMnb6n9RHNtc76Thn4wkWsCNWg6ayPMbwOWI=; _px2=eyJ1IjoiZWEwMmI2ZjAtZWNiYi0xMWVlLWI5NzktZTcyN2Q4YzdiYjUzIiwidiI6ImU0NzJlNjU1LWVjNDctMTFlZS1iNzViLTRjOGRmYTllM2YzYiIsInQiOjE3MTE2MDA1MzQzMDAsImgiOiIzMDdkMDkzNmIzYTQ5NmI3NDE1MWE1NWJmNGI3YzA3NzAwMGQ2N2RmOTZlZDdiZjMxZWRjNmNmZWY0YjZiMGZiIn0=; panoramaId_expiry=1711686634299; _pxde=f5b181c2348f2a522c515a4aaae1a1c759b0db9b0bed60d5e81ee489421b9d89:eyJ0aW1lc3RhbXAiOjE3MTE2MDAzMDUzNTYsImZfa2IiOjAsImlwY19pZCI6W119; _ga_GQ1PBLXZCT=GS1.1.1711600226.2.1.1711600323.0.0.0`
var userAgentString string = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36`

func TestChomeDp(t *testing.T) {

	// 创建带有自定义 User-Agent 的执行分配器
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgentString),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 1、创建谷歌浏览器实例
	ctx, cancel = chromedp.NewContext(
		context.Background(),
	)
	defer cancel()
	// 2、设置context 超时时间
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var siteTitle string
	var headlines []string

	cookies := splitCookies(cookieStr)
	// 执行任务
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, cookie := range cookies {
				err := network.SetCookie(cookie.Name, cookie.Value).
					WithDomain("www.bloomberg.com").Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		chromedp.Navigate(`https://www.bloomberg.com/ai`),
		chromedp.Title(&siteTitle), // 获取页面标题，以检查是否成功绕过了机器人检测
		// 使用适当的选择器来提取新闻标题
		chromedp.Evaluate(`Array.from(document.querySelectorAll('div[data-component="headline"] a')).map(a => a.textContent)`, &headlines),
	)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Site Title:", siteTitle)
	fmt.Println("Headlines:")
	for _, headline := range headlines {
		fmt.Println("-", headline)
	}
}

func TestParseEconomistList(t *testing.T) {
	p, err := proxy.RoundRobinProxySwitcher("http://127.0.0.1:8888")
	assert.Nil(t, err)

	f := spider.NewFetchService(spider.BrowserFetchType)
	//task.Logger = c.Logger
	rootreqs, err := BloombergTask.Rule.Root()
	BloombergTask.Proxy = p
	assert.Nil(t, err)

	var reqs []*spider.Request
	reqs = append(reqs, rootreqs...)

	for len(reqs) > 0 {

		req := reqs[0]
		reqs = reqs[1:]
		req.Task = BloombergTask
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
