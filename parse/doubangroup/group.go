package doubangroup

import (
	"fmt"
	"github.com/dreamerjackson/crawler/spider"
	"regexp"
)

const urlListRe = `(https://www.douban.com/group/topic/[0-9a-z]+/)"[^>]*>([^<]+)</a>`
const ContentRe = `<div class="topic-content">[\s\S]*?阳台[\s\S]*?<div class="aside">`

var DoubangroupTask = &spider.Task{
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
