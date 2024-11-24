package tasklib

import (
	"github.com/dreamerjackson/crawler/spider"
	"github.com/dreamerjackson/crawler/tasklib/bloomberg"
	"github.com/dreamerjackson/crawler/tasklib/doubanbook"
	"github.com/dreamerjackson/crawler/tasklib/doubangroup"
	"github.com/dreamerjackson/crawler/tasklib/doubangroupjs"
	"github.com/dreamerjackson/crawler/tasklib/economist"
	"github.com/dreamerjackson/crawler/tasklib/nytimes"
	"github.com/dreamerjackson/crawler/tasklib/semianalysis"
)

// https://github.com/twitterdev/Twitter-API-v2-sample-code/blob/8c63446fb6ed75b38283fca32e39c21ba08cd896/User-Lookup/get_users_me_user_context.py#L62

func init() {
	spider.TaskStore.Add(doubangroup.DoubangroupTask)
	spider.TaskStore.Add(doubanbook.DoubanBookTask)
	spider.TaskStore.Add(economist.EconomistTask)
	spider.TaskStore.Add(nytimes.NytimesTask)
	spider.TaskStore.Add(bloomberg.BloombergTask)
	spider.TaskStore.Add(semianalysis.Task)
	spider.TaskStore.AddJSTask(doubangroupjs.DoubangroupJSTask)
}
