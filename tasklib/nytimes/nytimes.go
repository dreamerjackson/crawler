package nytimes

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/spider"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"strings"
	"time"
)

var NytimesTask = &spider.Task{
	Options: spider.Options{
		Name: "nytimes_news",
		Limit: limiter.Multi(
			rate.NewLimiter(limiter.Per(1, 3*time.Second), 1),
			rate.NewLimiter(limiter.Per(20, 60*time.Second), 20),
		),
		Reload:   true,
		WaitTime: 2,
		MaxDepth: 5,
		Header: map[string]string{
			"Upgrade-Insecure-Requests": "1",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
		},
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					Priority: 1,
					URL:      "https://www.nytimes.com/spotlight/artificial-intelligence",
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
					"摘要",
					"作者",
				},
				ParseFunc: ParseNYTimesNews},
		},
	},
}

func ParseNYTimesNews(ctx *spider.Context) (spider.ParseResult, error) {
	//fmt.Println(string(ctx.Body))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(ctx.Body))
	if err != nil {
		ctx.Log.Error("goquery.NewDocumentFromReader failed",
			zap.Error(err))
		return spider.ParseResult{}, err
	}

	var items []*spider.DataCell
	// 遍历每个<li>元素
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		// 提取标题
		title := s.Find("h3").Text()

		// 提取链接，并补全为完整的URL
		link, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}
		completeLink := fmt.Sprintf("https://www.nytimes.com%s", strings.TrimSpace(link))

		// 提取摘要
		summary := s.Find("p").First().Text()

		// 提取作者信息
		authors := s.Find("article").Find("p").Last().Text()

		// 提取发布日期
		//date := s.Find("span[data-testid='todays-date']").Text()

		// 只有当必要的信息存在时才创建文章地图
		if title != "" && summary != "" {
			article := map[string]interface{}{
				"标题": title,
				"链接": completeLink,
				"摘要": summary,
				"作者": authors,
				//"发布日期": date,
			}

			data := ctx.Output(article)

			// 将文章添加到items切片中
			items = append(items, data)

			//fmt.Printf("标题: %s\n链接: %s\n摘要: %s\n作者: %s\n\n", title, completeLink, summary, authors)
		}
	})

	result := spider.ParseResult{
		Items: []interface{}{items},
	}

	return result, nil
}
