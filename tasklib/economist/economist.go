package economist

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

var EconomistTask = &spider.Task{
	Options: spider.Options{
		Name: "economist_list",
		Limit: limiter.Multi(
			rate.NewLimiter(limiter.Per(1, 3*time.Second), 1),
			rate.NewLimiter(limiter.Per(20, 60*time.Second), 20),
		),
		Reload:   true,
		WaitTime: 2,
		MaxDepth: 5,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					Priority: 1,
					URL:      "https://www.economist.com/",
					Method:   "GET",
					RuleName: "数据tag",
				},
			}

			return roots, nil
		},
		Trunk: map[string]*spider.Rule{
			"数据tag": {
				ItemFields: []string{
					"标题",
					"链接",
					"摘要",
				},
				ParseFunc: ParseTag},
		},
	},
}

func ParseTag(ctx *spider.Context) (spider.ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		ctx.Log.Error("goquery.NewDocumentFromReader failed: ",
			zap.Error(err))
	}
	var items []*spider.DataCell
	// 遍历所有包含<a>标签的<h3>
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		if s.ChildrenFiltered("a").Length() > 0 {
			title := s.Text()
			link, exists := s.Find("a").Attr("href")
			// 使用.NextAll().Filter("p").First()获取h3之后的第一个同级p标签
			summary := s.NextAll().Filter("p").First().Text()
			if exists && title != "" && summary != "" {
				// 确保链接是完整的
				completeLink := "https://www.economist.com" + strings.TrimSpace(link)

				// Create a map for each article
				article := map[string]interface{}{
					"标题": title,
					"链接": completeLink,
					"摘要": summary,
				}

				data := ctx.Output(article)

				// Add the article map to items slice
				items = append(items, data)

				// fmt.Printf("标题: %s\n链接: %s\n摘要: %s\n\n", title, completeLink, summary)
			}
		}
	})

	result := spider.ParseResult{
		Items: []interface{}{items},
	}

	return result, nil
}
