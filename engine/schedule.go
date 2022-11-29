package engine

import (
	"github.com/dreamerjackson/crawler/spider"
	"runtime/debug"
	"sync"

	"github.com/dreamerjackson/crawler/parse/doubanbook"
	"github.com/dreamerjackson/crawler/parse/doubangroup"
	"github.com/dreamerjackson/crawler/parse/doubangroupjs"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

func init() {
	Store.Add(doubangroup.DoubangroupTask)
	Store.Add(doubanbook.DoubanBookTask)
	Store.AddJSTask(doubangroupjs.DoubangroupJSTask)
}

func (c *CrawlerStore) Add(task *spider.Task) {
	c.Hash[task.Name] = task
	c.list = append(c.list, task)
}

// 用于动态规则添加请求。
func AddJsReqs(jreqs []map[string]interface{}) []*spider.Request {
	reqs := make([]*spider.Request, 0)

	for _, jreq := range jreqs {
		req := &spider.Request{}
		u, ok := jreq["URL"].(string)

		if !ok {
			return nil
		}

		req.URL = u
		req.RuleName, _ = jreq["RuleName"].(string)
		req.Method, _ = jreq["Method"].(string)
		req.Priority, _ = jreq["Priority"].(int64)
		reqs = append(reqs, req)
	}

	return reqs
}

// 用于动态规则添加请求。
func AddJsReq(jreq map[string]interface{}) []*spider.Request {
	reqs := make([]*spider.Request, 0)
	req := &spider.Request{}
	u, ok := jreq["URL"].(string)

	if !ok {
		return nil
	}

	req.URL = u
	req.RuleName, _ = jreq["RuleName"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Priority, _ = jreq["Priority"].(int64)
	reqs = append(reqs, req)

	return reqs
}

func (c *CrawlerStore) AddJSTask(m *spider.TaskModle) {
	task := &spider.Task{
		//Property: m.Property,
	}

	task.Rule.Root = func() ([]*spider.Request, error) {
		vm := otto.New()
		if err := vm.Set("AddJsReq", AddJsReqs); err != nil {
			return nil, err
		}

		v, err := vm.Eval(m.Root)

		if err != nil {
			return nil, err
		}

		e, err := v.Export()

		if err != nil {
			return nil, err
		}

		return e.([]*spider.Request), nil
	}

	for _, r := range m.Rules {
		paesrFunc := func(parse string) func(ctx *spider.Context) (spider.ParseResult, error) {
			return func(ctx *spider.Context) (spider.ParseResult, error) {
				vm := otto.New()
				if err := vm.Set("ctx", ctx); err != nil {
					return spider.ParseResult{}, err
				}

				v, err := vm.Eval(parse)

				if err != nil {
					return spider.ParseResult{}, err
				}

				e, err := v.Export()

				if err != nil {
					return spider.ParseResult{}, err
				}

				if e == nil {
					return spider.ParseResult{}, err
				}

				return e.(spider.ParseResult), err
			}
		}(r.ParseFunc)

		if task.Rule.Trunk == nil {
			task.Rule.Trunk = make(map[string]*spider.Rule, 0)
		}

		task.Rule.Trunk[r.Name] = &spider.Rule{
			ParseFunc: paesrFunc,
		}
	}

	c.Hash[task.Name] = task
	c.list = append(c.list, task)
}

// Store is a global instace
var Store = &CrawlerStore{
	list: []*spider.Task{},
	Hash: map[string]*spider.Task{},
}

func GetFields(taskName string, ruleName string) []string {
	return Store.Hash[taskName].Rule.Trunk[ruleName].ItemFields
}

type CrawlerStore struct {
	list []*spider.Task
	Hash map[string]*spider.Task
}

type Crawler struct {
	out         chan spider.ParseResult
	Visited     map[string]bool
	VisitedLock sync.Mutex

	failures    map[string]*spider.Request // 失败请求id -> 失败请求
	failureLock sync.Mutex

	options
}

type Scheduler interface {
	Schedule()
	Push(...*spider.Request)
	Pull() *spider.Request
}

type Schedule struct {
	requestCh   chan *spider.Request
	workerCh    chan *spider.Request
	priReqQueue []*spider.Request
	reqQueue    []*spider.Request
	Logger      *zap.Logger
}

func NewEngine(opts ...Option) *Crawler {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	e := &Crawler{}
	e.Visited = make(map[string]bool, 100)
	e.out = make(chan spider.ParseResult)
	e.failures = make(map[string]*spider.Request)
	e.options = options

	return e
}

func NewSchedule() *Schedule {
	s := &Schedule{}
	requestCh := make(chan *spider.Request)
	workerCh := make(chan *spider.Request)
	s.requestCh = requestCh
	s.workerCh = workerCh

	return s
}

func (c *Crawler) Run() {
	go c.Schedule()

	for i := 0; i < c.WorkCount; i++ {
		go c.CreateWork()
	}
	c.HandleResult()
}

func (s *Schedule) Push(reqs ...*spider.Request) {
	for _, req := range reqs {
		s.requestCh <- req
	}
}

func (s *Schedule) Pull() *spider.Request {
	r := <-s.workerCh

	return r
}

func (s *Schedule) Schedule() {
	var ch chan *spider.Request

	var req *spider.Request

	for {
		if req == nil && len(s.priReqQueue) > 0 {
			req = s.priReqQueue[0]
			s.priReqQueue = s.priReqQueue[1:]
			ch = s.workerCh
		}

		if req == nil && len(s.reqQueue) > 0 {
			req = s.reqQueue[0]
			s.reqQueue = s.reqQueue[1:]
			ch = s.workerCh
		}

		select {
		case r := <-s.requestCh:
			if r.Priority > 0 {
				s.priReqQueue = append(s.priReqQueue, r)
			} else {
				s.reqQueue = append(s.reqQueue, r)
			}
		case ch <- req:
			req = nil
			ch = nil
		}
	}
}

func (c *Crawler) Schedule() {
	var reqs []*spider.Request

	for _, task := range c.Seeds {
		t, ok := Store.Hash[task.Name]
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

	go c.scheduler.Schedule()
	go c.scheduler.Push(reqs...)
}

func (c *Crawler) CreateWork() {
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
			c.Logger.Error("check failed",
				zap.Error(err),
			)

			continue
		}

		if !req.Task.Reload && c.HasVisited(req) {
			c.Logger.Debug("request has visited",
				zap.String("url:", req.URL),
			)

			continue
		}

		c.StoreVisited(req)

		body, err := req.Fetch()
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

func (c *Crawler) HandleResult() {
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

func (c *Crawler) HasVisited(r *spider.Request) bool {
	c.VisitedLock.Lock()
	defer c.VisitedLock.Unlock()

	unique := r.Unique()

	return c.Visited[unique]
}

func (c *Crawler) StoreVisited(reqs ...*spider.Request) {
	c.VisitedLock.Lock()
	defer c.VisitedLock.Unlock()

	for _, r := range reqs {
		unique := r.Unique()
		c.Visited[unique] = true
	}
}

func (c *Crawler) SetFailure(req *spider.Request) {
	if !req.Task.Reload {
		c.VisitedLock.Lock()
		unique := req.Unique()
		delete(c.Visited, unique)
		c.VisitedLock.Unlock()
	}

	c.failureLock.Lock()
	defer c.failureLock.Unlock()

	if _, ok := c.failures[req.Unique()]; !ok {
		// 首次失败时，再重新执行一次
		c.failures[req.Unique()] = req
		c.scheduler.Push(req)
	}
}
