package doubangroupjs

import (
	"github.com/dreamerjackson/crawler/collect"
)

var DoubangroupJSTask = &collect.TaskModle{
	Property: collect.Property{
		Name:     "js_find_douban_sun_room",
		WaitTime: 2,
		MaxDepth: 5,
		Cookie:   "bid=-UXUw--yL5g; dbcl2=\"214281202:q0BBm9YC2Yg\"; __yadk_uid=jigAbrEOKiwgbAaLUt0G3yPsvehXcvrs; push_noty_num=0; push_doumail_num=0; __utmz=30149280.1665849857.1.1.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmv=30149280.21428; ck=SAvm; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1665925405%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.2072705865.1665849857.1665849857.1665925407.2; __utmc=30149280; __utmt=1; __utmb=30149280.23.5.1665925419338; _pk_id.100001.8cb4=fc1581490bf2b70c.1665849856.2.1665925421.1665849856.",
	},
	Root: `
		var arr = new Array();
 		for (var i = 25; i <= 25; i+=25) {
			var obj = {
			   Url: "https://www.douban.com/group/szsh/discussion?start=" + i,
			   Priority: 1,
			   RuleName: "解析网站URL",
			   Method: "GET",
		   };
			arr.push(obj);
		};
		console.log(arr[0].Url);
		AddJsReq(arr);
			`,
	Rules: []collect.RuleModle{
		{
			Name: "解析网站URL",
			ParseFunc: `
			ctx.ParseJSReg("解析阳台房","(https://www.douban.com/group/topic/[0-9a-z]+/)\"[^>]*>([^<]+)</a>");
			`,
		},
		{
			Name: "解析阳台房",
			ParseFunc: `
			//console.log("parse output");
			ctx.OutputJS("<div class=\"topic-content\">[\\s\\S]*?阳台[\\s\\S]*?<div class=\"aside\">");
			`,
		},
	},
}
