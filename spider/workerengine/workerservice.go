package workerengine

import (
	"github.com/dreamerjackson/crawler/spider"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"runtime/debug"
	"strings"
	"sync"
)

type WorkerService interface {
	Run(cluster bool)
	LoadResource() error
	WatchResource()
}

type workerService struct {
	out     chan spider.ParseResult
	rlock   sync.Mutex
	etcdCli *clientv3.Client
	options
}

func NewWorkerService(opts ...Option) (*workerService, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	e := &workerService{}
	e.out = make(chan spider.ParseResult)
	e.options = options

	// 任务加上默认的采集器与存储器
	for _, task := range spider.TaskStore.List {
		task.Fetcher = e.Fetcher
		task.Storage = e.Storage
	}

	return e, nil
}

func (c *workerService) Run(cluster bool) {
	if !cluster {
		c.handleSeeds()
	}
	c.LoadResource()
	go c.WatchResource()
	go c.scheduler.Schedule()
	for i := 0; i < c.WorkCount; i++ {
		go c.CreateWork()
	}
	c.HandleResult()
}

func (c *workerService) LoadResource() error {
	resources := make(map[string]*spider.ResourceSpec)
	resourceSpecs, err := c.resourceRegistry.GetResources()
	if err != nil {
		return err
	}

	for _, r := range resourceSpecs {
		id := getID(r.AssignedNode)
		if len(id) > 0 && c.id == id {
			resources[r.Name] = r
		}
	}

	c.Logger.Info("leader init load resource", zap.Int("lenth", len(resources)))
	c.resourceRepository.Set(resources)
	for _, r := range resources {
		c.runTasks(r.Name)
	}

	return nil
}

func (c *workerService) WatchResource() {
	watch := c.resourceRegistry.WatchResources()
	for w := range watch {
		if w.Canceled {
			c.Logger.Error("watch resource canceled")
			return
		}
		switch w.Typ {
		case spider.EventTypePut:
			c.rlock.Lock()
			if w.Res != nil {
				c.runTasks(w.Res.Name)
			}
			c.rlock.Unlock()
		case spider.EventTypeDelete:
			c.rlock.Lock()
			if w.Res != nil {
				c.deleteTasks(w.Res.Name)
			}
			c.rlock.Unlock()
		}
	}
}

func (c *workerService) deleteTasks(taskName string) {
	t, ok := spider.TaskStore.Hash[taskName]
	if !ok {
		c.Logger.Error("can not find preset tasks", zap.String("task name", taskName))
		return
	}
	t.Closed = true

	c.resourceRepository.Delete(taskName)
}

func (c *workerService) runTasks(name string) {

	if c.resourceRepository.HasResource(name) {
		c.Logger.Info("task has runing", zap.String("name", name))
		return
	}

	t, ok := spider.TaskStore.Hash[name]
	if !ok {
		c.Logger.Error("can not find preset tasks", zap.String("task name", name))
		return
	}
	t.Closed = false
	res, err := t.Rule.Root()

	if err != nil {
		c.Logger.Error("get root failed",
			zap.Error(err),
		)
		return
	}

	for _, req := range res {
		req.Task = t
	}

	c.scheduler.Push(res...)
}

func (c *workerService) handleSeeds() {
	var reqs []*spider.Request
	for _, task := range c.Seeds {
		t, ok := spider.TaskStore.Hash[task.Name]
		if !ok {
			c.Logger.Error("can not find preset tasks", zap.String("task name", task.Name))
			continue
		}
		task.Rule = t.Rule
		//task.Logger = c.Logger
		rootreqs, err := task.Rule.Root()

		if err != nil {
			c.Logger.Error("get root failed",
				zap.Error(err),
			)
			continue
		}

		for _, req := range rootreqs {
			req.Task = task
		}

		reqs = append(reqs, rootreqs...)
	}
	go c.scheduler.Push(reqs...)
}

func (c *workerService) CreateWork() {
	defer func() {
		if err := recover(); err != nil {
			c.Logger.Error("worker panic",
				zap.Any("err", err),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	for {
		req := c.scheduler.Pull()
		if err := req.Check(); err != nil {
			c.Logger.Debug("check failed",
				zap.Error(err),
			)

			continue
		}

		if !req.Task.Reload && c.reqRepository.HasVisited(req) {
			c.Logger.Debug("request has visited",
				zap.String("url:", req.URL),
			)

			continue
		}

		c.reqRepository.AddVisited(req)

		body, err := req.Task.Fetcher.Get(req)
		if err != nil {
			c.Logger.Error("can't fetch ",
				zap.Error(err),
				zap.String("url", req.URL),
			)
			c.SetFailure(req)

			continue
		}

		if len(body) < 6000 {
			c.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", req.URL),
			)
			c.SetFailure(req)

			continue
		}

		rule := req.Task.Rule.Trunk[req.RuleName]
		ctx := &spider.Context{
			Body: body,
			Req:  req,
		}
		result, err := rule.ParseFunc(ctx)

		if err != nil {
			c.Logger.Error("ParseFunc failed ",
				zap.Error(err),
				zap.String("url", req.URL),
			)

			continue
		}

		if len(result.Requesrts) > 0 {
			go c.scheduler.Push(result.Requesrts...)
		}

		c.out <- result
	}
}

func (c *workerService) HandleResult() {
	for result := range c.out {
		for _, item := range result.Items {
			switch d := item.(type) {
			case *spider.DataCell:
				if err := d.Task.Storage.Save(d); err != nil {
					c.Logger.Error("")
				}
			}
			c.Logger.Sugar().Info("get result: ", item)
		}
	}
}

func (c *workerService) SetFailure(req *spider.Request) {
	if !req.Task.Reload {
		c.reqRepository.DeleteVisited(req)
	}

	if !c.reqRepository.AddFailures(req) {
		// 首次失败时，再重新执行一次
		c.scheduler.Push(req)
	}
}

func getID(assignedNode string) string {
	s := strings.Split(assignedNode, "|")
	if len(s) < 2 {
		return ""
	}
	return s[0]
}
