package doubangroup

import (
	"fmt"
	"github.com/dreamerjackson/crawler/spider"
	"regexp"
)

const urlListRe = `(https://www.douban.com/group/topic/[0-9a-z]+/)"[^>]*>([^<]+)</a>`
const ContentRe = `<div class="topic-content">[\s\S]*?阳台[\s\S]*?<div class="aside">`

var DoubangroupTask = &spider.Task{
	//Property: spider.Property{
	//	Name:     "find_douban_sun_room",
	//	WaitTime: 2,
	//	MaxDepth: 5,
	//	Cookie:   "bid=-UXUw--yL5g; dbcl2=\"214281202:q0BBm9YC2Yg\"; __yadk_uid=jigAbrEOKiwgbAaLUt0G3yPsvehXcvrs; push_noty_num=0; push_doumail_num=0; __utmz=30149280.1665849857.1.1.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmv=30149280.21428; ck=SAvm; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1665925405%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.2072705865.1665849857.1665849857.1665925407.2; __utmc=30149280; __utmt=1; __utmb=30149280.23.5.1665925419338; _pk_id.100001.8cb4=fc1581490bf2b70c.1665849856.2.1665925421.1665849856.",
	//},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			var roots []*spider.Request
			for i := 0; i < 25; i += 25 {
				str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
				roots = append(roots, &spider.Request{
					Priority: 1,
					URL:      str,
					Method:   "GET",
					RuleName: "解析网站URL",
				})
			}

			return roots, nil
		},
		Trunk: map[string]*spider.Rule{
			"解析网站URL": {ParseFunc: ParseURL},
			"解析阳台房":   {ParseFunc: GetSunRoom},
		},
	},
}

func ParseURL(ctx *spider.Context) (spider.ParseResult, error) {
	re := regexp.MustCompile(urlListRe)

	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := spider.ParseResult{}

	for _, m := range matches {
		u := string(m[1])

		result.Requesrts = append(
			result.Requesrts, &spider.Request{
				Method:   "GET",
				Task:     ctx.Req.Task,
				URL:      u,
				Depth:    ctx.Req.Depth + 1,
				RuleName: "解析阳台房",
			})
	}

	return result, nil
}

func GetSunRoom(ctx *spider.Context) (spider.ParseResult, error) {
	re := regexp.MustCompile(ContentRe)

	if ok := re.Match(ctx.Body); !ok {
		return spider.ParseResult{
			Items: []interface{}{},
		}, nil
	}

	result := spider.ParseResult{
		Items: []interface{}{ctx.Req.URL},
	}

	return result, nil
}
