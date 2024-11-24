package bloomberg

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/spider"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"strings"
	"time"
)

var BloombergTask = &spider.Task{
	Options: spider.Options{
		Name: "bloomberg_ai_news",
		Limit: limiter.Multi(
			rate.NewLimiter(limiter.Per(1, 3*time.Second), 1),
			rate.NewLimiter(limiter.Per(20, 60*time.Second), 20),
		),
		Reload:   true,
		WaitTime: 2,
		MaxDepth: 5,
		Cookie:   `optimizelyEndUserId=oeu1711550399176r0.1877745177569461; bbgconsentstring=req1fun1pad1; bdfpc=004.9447305864.1711550400529; _gcl_au=1.1.1799522145.1711550401; consentUUID=c858b581-08ff-4283-bce4-67f4d9b469ba; _fbp=fb.1.1711550401204.1932059742; usnatUUID=f2ace540-bc32-45fa-9cd5-380fd71e788a; pxcts=e472f5df-ec47-11ee-b75b-1b6d6af0f63f; _pxvid=e472e655-ec47-11ee-b75b-4c8dfa9e3f3b; agent_id=1f931978-32eb-4c2b-adeb-56dabf0ee4ab; session_id=a15a91ba-c52e-4150-9967-ec9cc2b88264; session_key=9ad6d5dad31de9a68877b65a95b3d3ba0adca711; gatehouse_id=7c24689a-e9aa-4e70-8b65-5916faa74d28; geo_info=%7B%22countryCode%22%3A%22HK%22%2C%22country%22%3A%22HK%22%2C%22field_n%22%3A%22cp%22%2C%22trackingRegion%22%3A%22Asia%22%2C%22cacheExpiredTime%22%3A1712155201555%2C%22region%22%3A%22Asia%22%2C%22fieldN%22%3A%22cp%22%7D%7C1712155201555; exp_pref=APAC; country_code=HK; seen_uk=1; _gcl_aw=GCL.1711550412.CjwKCAjwh4-wBhB3EiwAeJsppNyuWpdTxz82Fs7ec8vFx5f8BAo40NNfkuhoEz1ePJJOF_yLTjtWLhoCGD4QAvD_BwE; _gcl_dc=GCL.1711550412.CjwKCAjwh4-wBhB3EiwAeJsppNyuWpdTxz82Fs7ec8vFx5f8BAo40NNfkuhoEz1ePJJOF_yLTjtWLhoCGD4QAvD_BwE; afUserId=f1d90c08-d5d5-4750-a181-58830064d649-p; AF_SYNC=1711550414351; _cc_id=4b00edd3be9c419a4ca180d3533f94f5; panoramaId=1fd98f3669d105fc90dc837acd15a9fb927ae6090142d1edb7c3b9439c610e4e; panoramaIdType=panoDevice; _scid=6385e024-98dc-4487-87b5-3e2da7c1b415; _sctr=1%7C1711468800000; __stripe_mid=7905422e-2c3f-4d4d-b502-cac5a0831e8be7dcb2; _scid_r=6385e024-98dc-4487-87b5-3e2da7c1b415; professional-cookieConsent=new-relic|perimeterx-bot-detection|perimeterx-pixel|google-tag-manager|google-analytics|microsoft-clarity|optimizely|microsoft-advertising|eloqua|adwords|linkedin-insights; _gid=GA1.2.267148422.1711556267; drift_aid=5f79dcc1-c67f-413f-bd72-f8397766c4cc; driftt_aid=5f79dcc1-c67f-413f-bd72-f8397766c4cc; _ga_NNP7N7T2TG=GS1.1.1711556267.1.1.1711556279.48.0.0; _ga=GA1.1.928448926.1711550401; __gads=ID=0adf99e208745bfb:T=1711550412:RT=1711600226:S=ALNI_MaftDuDZJA8HYObZ08pS8lwAPixcQ; __gpi=UID=00000d6ad874281c:T=1711550412:RT=1711600226:S=ALNI_MbaCozYcJsrtQwyEUxfIXAgIQN0CQ; __eoi=ID=39c1f3bd6f4a61b5:T=1711550412:RT=1711600226:S=AA-AfjZL1Az9HIYk5RZl4ilz7r9K; _clck=8d76t0%7C2%7Cfkg%7C0%7C1547; _parsely_session={%22sid%22:2%2C%22surl%22:%22https://www.bloomberg.com/asia%22%2C%22sref%22:%22%22%2C%22sts%22:1711600227224%2C%22slts%22:1711550401154}; _parsely_visitor={%22id%22:%22pid=7f1bee4ee7907a07f60ca606849702d6%22%2C%22session_count%22:2%2C%22last_session_ts%22:1711600227224}; _user-data=%7B%22status%22%3A%22anonymous%22%7D; __sppvid=36e9b221-932f-40cf-bc63-2f4007873f35; _uetsid=e419b980ec4711eea00277baae156ef6; _uetvid=fa9a1f20455411ec9b74dbfc533fddae; _reg-csrf-token=dWw8H8xh-YFIH7yxAWwE656GiBSM_5LJhDMc; _last-refresh=2024-3-28%204%3A30; _clsk=ro46fm%7C1711600234083%7C2%7C0%7Cn.clarity.ms%2Fcollect; _px3=3a49a8cecb63d99751073132d1f058c98a70f771eff02b1ce92cc76fcd979975:ToMyJqXIe4ITefCtC3LC7xJifq/LgdtbAKBF6Lj3fnt1u/hZMiyy+og+/r1V1CoE7q296ur52jJT0hFqvWwymw==:1000:YDyhkzul5GXL8980W1G1bH5GEVYWR58ma+myGto7jBcxh/ZOurLACbpahMPjXoD3YxYdO/KQ67rpzuoAy/5Tf/q0PVESof7jVCicrZyLhZSXXMq372zMRdMnPl4jqu4x+kLvfHPPw8kdN7FVcaD1Rb1fzWFDZaQKPGAi2BEpksWqrrVfa/AXFNpLpTRJhVYWxQ04hNBtMnb6n9RHNtc76Thn4wkWsCNWg6ayPMbwOWI=; _px2=eyJ1IjoiZWEwMmI2ZjAtZWNiYi0xMWVlLWI5NzktZTcyN2Q4YzdiYjUzIiwidiI6ImU0NzJlNjU1LWVjNDctMTFlZS1iNzViLTRjOGRmYTllM2YzYiIsInQiOjE3MTE2MDA1MzQzMDAsImgiOiIzMDdkMDkzNmIzYTQ5NmI3NDE1MWE1NWJmNGI3YzA3NzAwMGQ2N2RmOTZlZDdiZjMxZWRjNmNmZWY0YjZiMGZiIn0=; panoramaId_expiry=1711686634299; _pxde=f5b181c2348f2a522c515a4aaae1a1c759b0db9b0bed60d5e81ee489421b9d89:eyJ0aW1lc3RhbXAiOjE3MTE2MDAzMDUzNTYsImZfa2IiOjAsImlwY19pZCI6W119; _ga_GQ1PBLXZCT=GS1.1.1711600226.2.1.1711600323.0.0.0`,
		Header: map[string]string{
			"Host":                      "www.bloomberg.com",
			"Upgrade-Insecure-Requests": "1",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
			"Cookie":                    `optimizelyEndUserId=oeu1711550399176r0.1877745177569461; bbgconsentstring=req1fun1pad1; bdfpc=004.9447305864.1711550400529; _gcl_au=1.1.1799522145.1711550401; consentUUID=c858b581-08ff-4283-bce4-67f4d9b469ba; _fbp=fb.1.1711550401204.1932059742; usnatUUID=f2ace540-bc32-45fa-9cd5-380fd71e788a; pxcts=e472f5df-ec47-11ee-b75b-1b6d6af0f63f; _pxvid=e472e655-ec47-11ee-b75b-4c8dfa9e3f3b; agent_id=1f931978-32eb-4c2b-adeb-56dabf0ee4ab; session_id=a15a91ba-c52e-4150-9967-ec9cc2b88264; session_key=9ad6d5dad31de9a68877b65a95b3d3ba0adca711; gatehouse_id=7c24689a-e9aa-4e70-8b65-5916faa74d28; geo_info=%7B%22countryCode%22%3A%22HK%22%2C%22country%22%3A%22HK%22%2C%22field_n%22%3A%22cp%22%2C%22trackingRegion%22%3A%22Asia%22%2C%22cacheExpiredTime%22%3A1712155201555%2C%22region%22%3A%22Asia%22%2C%22fieldN%22%3A%22cp%22%7D%7C1712155201555; exp_pref=APAC; country_code=HK; seen_uk=1; _gcl_aw=GCL.1711550412.CjwKCAjwh4-wBhB3EiwAeJsppNyuWpdTxz82Fs7ec8vFx5f8BAo40NNfkuhoEz1ePJJOF_yLTjtWLhoCGD4QAvD_BwE; _gcl_dc=GCL.1711550412.CjwKCAjwh4-wBhB3EiwAeJsppNyuWpdTxz82Fs7ec8vFx5f8BAo40NNfkuhoEz1ePJJOF_yLTjtWLhoCGD4QAvD_BwE; afUserId=f1d90c08-d5d5-4750-a181-58830064d649-p; AF_SYNC=1711550414351; _cc_id=4b00edd3be9c419a4ca180d3533f94f5; panoramaId=1fd98f3669d105fc90dc837acd15a9fb927ae6090142d1edb7c3b9439c610e4e; panoramaIdType=panoDevice; _scid=6385e024-98dc-4487-87b5-3e2da7c1b415; _sctr=1%7C1711468800000; __stripe_mid=7905422e-2c3f-4d4d-b502-cac5a0831e8be7dcb2; _scid_r=6385e024-98dc-4487-87b5-3e2da7c1b415; professional-cookieConsent=new-relic|perimeterx-bot-detection|perimeterx-pixel|google-tag-manager|google-analytics|microsoft-clarity|optimizely|microsoft-advertising|eloqua|adwords|linkedin-insights; _gid=GA1.2.267148422.1711556267; drift_aid=5f79dcc1-c67f-413f-bd72-f8397766c4cc; driftt_aid=5f79dcc1-c67f-413f-bd72-f8397766c4cc; _ga_NNP7N7T2TG=GS1.1.1711556267.1.1.1711556279.48.0.0; _ga=GA1.1.928448926.1711550401; __gads=ID=0adf99e208745bfb:T=1711550412:RT=1711600226:S=ALNI_MaftDuDZJA8HYObZ08pS8lwAPixcQ; __gpi=UID=00000d6ad874281c:T=1711550412:RT=1711600226:S=ALNI_MbaCozYcJsrtQwyEUxfIXAgIQN0CQ; __eoi=ID=39c1f3bd6f4a61b5:T=1711550412:RT=1711600226:S=AA-AfjZL1Az9HIYk5RZl4ilz7r9K; _clck=8d76t0%7C2%7Cfkg%7C0%7C1547; _parsely_session={%22sid%22:2%2C%22surl%22:%22https://www.bloomberg.com/asia%22%2C%22sref%22:%22%22%2C%22sts%22:1711600227224%2C%22slts%22:1711550401154}; _parsely_visitor={%22id%22:%22pid=7f1bee4ee7907a07f60ca606849702d6%22%2C%22session_count%22:2%2C%22last_session_ts%22:1711600227224}; _user-data=%7B%22status%22%3A%22anonymous%22%7D; __sppvid=36e9b221-932f-40cf-bc63-2f4007873f35; _uetsid=e419b980ec4711eea00277baae156ef6; _uetvid=fa9a1f20455411ec9b74dbfc533fddae; _reg-csrf-token=dWw8H8xh-YFIH7yxAWwE656GiBSM_5LJhDMc; _last-refresh=2024-3-28%204%3A30; _clsk=ro46fm%7C1711600234083%7C2%7C0%7Cn.clarity.ms%2Fcollect; _px3=3a49a8cecb63d99751073132d1f058c98a70f771eff02b1ce92cc76fcd979975:ToMyJqXIe4ITefCtC3LC7xJifq/LgdtbAKBF6Lj3fnt1u/hZMiyy+og+/r1V1CoE7q296ur52jJT0hFqvWwymw==:1000:YDyhkzul5GXL8980W1G1bH5GEVYWR58ma+myGto7jBcxh/ZOurLACbpahMPjXoD3YxYdO/KQ67rpzuoAy/5Tf/q0PVESof7jVCicrZyLhZSXXMq372zMRdMnPl4jqu4x+kLvfHPPw8kdN7FVcaD1Rb1fzWFDZaQKPGAi2BEpksWqrrVfa/AXFNpLpTRJhVYWxQ04hNBtMnb6n9RHNtc76Thn4wkWsCNWg6ayPMbwOWI=; _px2=eyJ1IjoiZWEwMmI2ZjAtZWNiYi0xMWVlLWI5NzktZTcyN2Q4YzdiYjUzIiwidiI6ImU0NzJlNjU1LWVjNDctMTFlZS1iNzViLTRjOGRmYTllM2YzYiIsInQiOjE3MTE2MDA1MzQzMDAsImgiOiIzMDdkMDkzNmIzYTQ5NmI3NDE1MWE1NWJmNGI3YzA3NzAwMGQ2N2RmOTZlZDdiZjMxZWRjNmNmZWY0YjZiMGZiIn0=; panoramaId_expiry=1711686634299; _pxde=f5b181c2348f2a522c515a4aaae1a1c759b0db9b0bed60d5e81ee489421b9d89:eyJ0aW1lc3RhbXAiOjE3MTE2MDAzMDUzNTYsImZfa2IiOjAsImlwY19pZCI6W119; _ga_GQ1PBLXZCT=GS1.1.1711600226.2.1.1711600323.0.0.0`,
		},
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					Priority: 1,
					URL:      "https://www.bloomberg.com/ai",
					Method:   "GET",
					RuleName: "headline",
				},
			}

			return roots, nil
		},
		Trunk: map[string]*spider.Rule{
			"headline": {
				ItemFields: []string{
					"标题",
					"链接",
				},
				ParseFunc: ParseBloombergHeadline},
		},
	},
}

func ParseBloombergHeadline(ctx *spider.Context) (spider.ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		ctx.Log.Error("goquery.NewDocumentFromReader failed: ",
			zap.Error(err))
	}
	var items []*spider.DataCell

	// 选择器使用 div[data-component="headline"] 来找到新闻标题
	doc.Find(`div[data-component="headline"] > a`).Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		link, exists := s.Attr("href")
		if exists && title != "" && len(title) > 10 {
			// 确保链接是完整的
			completeLink := strings.TrimSpace(link)
			// 打印标题和链接
			//fmt.Printf("标题: %s\n 链接: %s \n\n", title, completeLink)
			// Create a map for each news article
			article := map[string]interface{}{
				"标题": title,
				"链接": completeLink,
			}

			data := ctx.Output(article)

			// Add the article map to items slice
			items = append(items, data)
		}
	})

	result := spider.ParseResult{}

	if len(items) > 0 {
		result.Items = []interface{}{items}
	}

	return result, nil
}
