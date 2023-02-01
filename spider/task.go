package spider

import (
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/proxy"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 一个任务实例，
type Task struct {
	Visited     map[string]bool
	VisitedLock sync.Mutex
	//
	Closed bool
	Rule   RuleTree
	Options
}

type TaskConfig struct {
	Name     string
	Cookie   string
	WaitTime int64
	Reload   bool
	MaxDepth int64
	Fetcher  string
	Limits   []LimitCofig
}

type LimitCofig struct {
	EventCount int
	EventDur   int // 秒
	Bucket     int // 桶大小
}

func NewTask(opts ...Option) *Task {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	d := &Task{}
	d.Options = options

	return d
}

type Options struct {
	Name     string        `json:"name"` // 任务名称，应保证唯一性
	URL      string        `json:"url"`
	Cookie   string        `json:"cookie"`
	WaitTime int64         `json:"wait_time"` // 随机休眠时间，秒
	Reload   bool          `json:"reload"`    // 网站是否可以重复爬取
	MaxDepth int64         `json:"max_depth"`
	Timeout  time.Duration // http超时时间
	Proxy    proxy.Func
	Fetcher  Fetcher
	Storage  DataRepository
	Limit    limiter.RateLimiter
	logger   *zap.Logger
}

var defaultOptions = Options{
	logger:   zap.NewNop(),
	WaitTime: 5,
	Reload:   false,
	MaxDepth: 5,
	Timeout:  3 * time.Second,
}

type Option func(opts *Options)

func WithLogger(logger *zap.Logger) Option {
	return func(opts *Options) {
		opts.logger = logger
	}
}

func WithName(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func WithURL(url string) Option {
	return func(opts *Options) {
		opts.URL = url
	}
}

func WithCookie(cookie string) Option {
	return func(opts *Options) {
		opts.Cookie = cookie
	}
}

func WithWaitTime(waitTime int64) Option {
	return func(opts *Options) {
		opts.WaitTime = waitTime
	}
}

func WithReload(reload bool) Option {
	return func(opts *Options) {
		opts.Reload = reload
	}
}

func WithFetcher(f Fetcher) Option {
	return func(opts *Options) {
		opts.Fetcher = f
	}
}

func WithStorage(s DataRepository) Option {
	return func(opts *Options) {
		opts.Storage = s
	}
}

func WithMaxDepth(maxDepth int64) Option {
	return func(opts *Options) {
		opts.MaxDepth = maxDepth
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

func WithProxy(proxy proxy.Func) Option {
	return func(opts *Options) {
		opts.Proxy = proxy
	}
}
