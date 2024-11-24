package semianalysis

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/spider"
	"golang.org/x/time/rate"
	"strings"
	"time"
)

var Task = &spider.Task{
	Options: spider.Options{
		Name: "semianalysis",
		Limit: limiter.Multi(
			rate.NewLimiter(limiter.Per(1, 3*time.Second), 1),
			rate.NewLimiter(limiter.Per(20, 60*time.Second), 20),
		),
		Reload:   true,
		WaitTime: 2,
		MaxDepth: 5,
		Header: map[string]string{
			"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
		},
		Fetcher: nil,
	},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					Priority: 1,
					URL:      "https://www.semianalysis.com/?sort=new",
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
					"发布日期",
				},
				ParseFunc: ParseNYTimesNews},
		},
	},
}

func ParseNYTimesNews(ctxrr *spider.Context) (spider.ParseResult, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("http://127.0.0.1:8888"),
	)
	// 创建上下文并分配资源
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a chromedp context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set a timeout
	ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	// Enable networking
	err := chromedp.Run(ctx, network.Enable())
	if err != nil {
		panic(err)
	}
	var pageHTML string

	//var articleData []map[string]string
	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.semianalysis.com/`),
		chromedp.WaitVisible(`body`, chromedp.ByQuery), // 确保<body>标签可见，确保页面已加载
		chromedp.OuterHTML(`html`, &pageHTML),          // 获取整个<html>标签的外部HTML
	)

	//fmt.Println("Page HTML:")
	fmt.Println("bodylength:", len(pageHTML)) // 打印获取到的HTML，可以查看整个页面的HTML结构
	if err != nil {
		return spider.ParseResult{}, err
	}

	// Process the HTML using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageHTML))
	if err != nil {
		return spider.ParseResult{}, err
	}

	var items []*spider.DataCell

	// Using more specific selectors to find articles within elements with the 'pencraft' class
	doc.Find("div.pencraft a").Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		link, exists := s.Attr("href")

		if !exists {
			return
		}
		date := s.Find("time").Text() // Assumes date is in a <time> tag within .pencraft

		if date == "" {
			// If not found, try finding it in the nearest parent or sibling
			date = s.Parents().Find("time").First().Text()
			if date == "" {
				date = s.Siblings().Find("time").First().Text()
			}
		}

		if strings.Contains(link, "semianalysis.com") && title != "" && len(title) > 10 && date != "" {
			article := map[string]interface{}{
				"标题":   title,
				"链接":   link,
				"发布日期": date,
			}
			data := ctxrr.Output(article)
			// 将文章添加到items切片中
			items = append(items, data)
		}
	})

	result := spider.ParseResult{
		Items: []interface{}{items},
	}

	return result, nil
}
