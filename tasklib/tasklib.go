package tasklib

import (
	"github.com/dreamerjackson/crawler/spider"
	"github.com/dreamerjackson/crawler/tasklib/doubanbook"
	"github.com/dreamerjackson/crawler/tasklib/doubangroup"
	"github.com/dreamerjackson/crawler/tasklib/doubangroupjs"
)

func init() {
	spider.TaskStore.Add(doubangroup.DoubangroupTask)
	spider.TaskStore.Add(doubanbook.DoubanBookTask)
	spider.TaskStore.AddJSTask(doubangroupjs.DoubangroupJSTask)
}
