package doubangroup

import (
	"fmt"
	"github.com/dreamerjackson/crawler/collect"
	"regexp"
	"time"
)

const urlListRe = `(https://www.douban.com/group/topic/[0-9a-z]+/)"[^>]*>([^<]+)</a>`
const ContentRe = `<div class="topic-content">[\s\S]*?阳台[\s\S]*?<div class="aside">`

var DoubangroupTask = &collect.Task{
	Name:     "find_douban_sun_room",
	WaitTime: 1 * time.Second,
	MaxDepth: 5,
	Cookie:   "bid=-UXUw--yL5g; dbcl2=\"214281202:q0BBm9YC2Yg\"; __yadk_uid=jigAbrEOKiwgbAaLUt0G3yPsvehXcvrs; push_noty_num=0; push_doumail_num=0; __utmz=30149280.1665849857.1.1.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmv=30149280.21428; ck=SAvm; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1665925405%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.2072705865.1665849857.1665849857.1665925407.2; __utmc=30149280; __utmt=1; __utmb=30149280.23.5.1665925419338; _pk_id.100001.8cb4=fc1581490bf2b70c.1665849856.2.1665925421.1665849856.",
	Rule: collect.RuleTree{
		Root: func() []*collect.Request {
			var roots []*collect.Request
			for i := 0; i < 25; i += 25 {
				str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
				roots = append(roots, &collect.Request{
					Priority: 1,
					Url:      str,
					Method:   "GET",
					RuleName: "解析网站URL",
				})
			}
			return roots
		},
		Trunk: map[string]*collect.Rule{
			"解析网站URL": &collect.Rule{ParseURL},
			"解析阳台房":   &collect.Rule{GetSunRoom},
		},
	},
}

func ParseURL(ctx *collect.Context) collect.ParseResult {
	re := regexp.MustCompile(urlListRe)

	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}

	for _, m := range matches {
		u := string(m[1])
		result.Requesrts = append(
			result.Requesrts, &collect.Request{
				Method:   "GET",
				Task:     ctx.Req.Task,
				Url:      u,
				Depth:    ctx.Req.Depth + 1,
				RuleName: "解析阳台房",
			})
	}
	return result
}

func GetSunRoom(ctx *collect.Context) collect.ParseResult {
	re := regexp.MustCompile(ContentRe)

	ok := re.Match(ctx.Body)
	if !ok {
		return collect.ParseResult{
			Items: []interface{}{},
		}
	}
	result := collect.ParseResult{
		Items: []interface{}{ctx.Req.Url},
	}
	return result
}
