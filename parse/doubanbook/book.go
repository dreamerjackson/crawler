package doubanbook

import (
	"github.com/dreamerjackson/crawler/collect"
	"regexp"
	"strconv"
	"time"
)

var DoubanBookTask = &collect.Task{
	Property: collect.Property{
		Name:     "douban_book_list",
		WaitTime: 1 * time.Second,
		MaxDepth: 5,
		Cookie:   "bid=-UXUw--yL5g; push_doumail_num=0; __utmv=30149280.21428; __utmc=30149280; __gads=ID=c6eaa3cb04d5733a-2259490c18d700e1:T=1666111347:RT=1666111347:S=ALNI_MaonVB4VhlZG_Jt25QAgq-17DGDfw; frodotk_db=\"17dfad2f83084953479f078e8918dbf9\"; gr_user_id=cecf9a7f-2a69-4dfd-8514-343ca5c61fb7; __utmc=81379588; _vwo_uuid_v2=D55C74107BD58A95BEAED8D4E5B300035|b51e2076f12dc7b2c24da50b77ab3ffe; __yadk_uid=BKBuETKRjc2fmw3QZuSw4rigUGsRR4wV; ct=y; ll=\"108288\"; viewed=\"36104107\"; ap_v=0,6.0; __gpi=UID=000008887412003e:T=1666111347:RT=1668851750:S=ALNI_MZmNsuRnBrad4_ynFUhTl0Hi0l5oA; __utma=30149280.2072705865.1665849857.1668851747.1668854335.25; __utmz=30149280.1668854335.25.4.utmcsr=douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/misc/sorry; __utma=81379588.990530987.1667661846.1668852024.1668854335.8; __utmz=81379588.1668854335.8.2.utmcsr=douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/misc/sorry; _pk_ref.100001.3ac3=[\"\",\"\",1668854335,\"https://www.douban.com/misc/sorry?original-url=https%3A%2F%2Fbook.douban.com%2Ftag%2F%25E5%25B0%258F%25E8%25AF%25B4\"]; _pk_ses.100001.3ac3=*; gr_cs1_5f43ac5c-3e30-4ffd-af0e-7cd5aadeb3d1=user_id:0; __utmt=1; dbcl2=\"214281202:GLkwnNqtJa8\"; ck=dBZD; gr_session_id_22c937bbd8ebd703f2d8e9445f7dfd03=ca04de17-2cbf-4e45-914a-428d3c26cfe3; gr_cs1_ca04de17-2cbf-4e45-914a-428d3c26cfe3=user_id:1; __utmt_douban=1; gr_session_id_22c937bbd8ebd703f2d8e9445f7dfd03_ca04de17-2cbf-4e45-914a-428d3c26cfe3=true; __utmb=30149280.10.10.1668854335; __utmb=81379588.9.10.1668854335; _pk_id.100001.3ac3=02339dd9cc7d293a.1667661846.8.1668855011.1668852362.; push_noty_num=0",
	},
	Rule: collect.RuleTree{
		Root: func() ([]*collect.Request, error) {
			roots := []*collect.Request{
				&collect.Request{
					Priority: 1,
					Url:      "https://book.douban.com",
					Method:   "GET",
					RuleName: "数据tag",
				},
			}
			return roots, nil
		},
		Trunk: map[string]*collect.Rule{
			"数据tag": &collect.Rule{ParseFunc: ParseTag},
			"书籍列表":  &collect.Rule{ParseFunc: ParseBookList},
			"书籍简介": &collect.Rule{
				ItemFields: []string{
					"书名",
					"作者",
					"页数",
					"出版社",
					"得分",
					"价格",
					"简介",
				},
				ParseFunc: ParseBookDetail,
			},
		},
	},
}

const regexpStr = `<a href="([^"]+)" class="tag">([^<]+)</a>`

func ParseTag(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(regexpStr)

	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}

	for _, m := range matches {
		result.Requesrts = append(
			result.Requesrts, &collect.Request{
				Method:   "GET",
				Task:     ctx.Req.Task,
				Url:      "https://book.douban.com" + string(m[1]),
				Depth:    ctx.Req.Depth + 1,
				RuleName: "书籍列表",
			})
	}
	// 在添加limit之前，临时减少抓取数量,防止被服务器封禁
	result.Requesrts = result.Requesrts[:1]
	return result, nil
}

const BooklistRe = `<a.*?href="([^"]+)" title="([^"]+)"`

func ParseBookList(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(BooklistRe)
	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}
	for _, m := range matches {
		req := &collect.Request{
			Method:   "GET",
			Task:     ctx.Req.Task,
			Url:      string(m[1]),
			Depth:    ctx.Req.Depth + 1,
			RuleName: "书籍简介",
		}
		req.TmpData = &collect.Temp{}
		req.TmpData.Set("book_name", string(m[2]))
		result.Requesrts = append(result.Requesrts, req)
	}
	// 在添加limit之前，临时减少抓取数量,防止被服务器封禁
	result.Requesrts = result.Requesrts[:3]

	return result, nil
}

var autoRe = regexp.MustCompile(`<span class="pl"> 作者</span>:[\d\D]*?<a.*?>([^<]+)</a>`)
var public = regexp.MustCompile(`<span class="pl">出版社:</span>([^<]+)<br/>`)
var pageRe = regexp.MustCompile(`<span class="pl">页数:</span> ([^<]+)<br/>`)
var priceRe = regexp.MustCompile(`<span class="pl">定价:</span>([^<]+)<br/>`)
var scoreRe = regexp.MustCompile(`<strong class="ll rating_num " property="v:average">([^<]+)</strong>`)
var intoRe = regexp.MustCompile(`<div class="intro">[\d\D]*?<p>([^<]+)</p></div>`)

func ParseBookDetail(ctx *collect.Context) (collect.ParseResult, error) {
	bookName := ctx.Req.TmpData.Get("book_name")
	page, _ := strconv.Atoi(ExtraString(ctx.Body, pageRe))

	book := map[string]interface{}{
		"书名":  bookName,
		"作者":  ExtraString(ctx.Body, autoRe),
		"页数":  page,
		"出版社": ExtraString(ctx.Body, public),
		"得分":  ExtraString(ctx.Body, scoreRe),
		"价格":  ExtraString(ctx.Body, priceRe),
		"简介":  ExtraString(ctx.Body, intoRe),
	}
	data := ctx.Output(book)

	result := collect.ParseResult{
		Items: []interface{}{data},
	}

	return result, nil
}

func ExtraString(contents []byte, re *regexp.Regexp) string {

	match := re.FindSubmatch(contents)

	if len(match) >= 2 {
		return string(match[1])
	} else {
		return ""
	}
}
