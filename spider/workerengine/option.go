package workerengine

import (
	"github.com/dreamerjackson/crawler/spider"
	"go.uber.org/zap"
)

type Option func(opts *options)

type options struct {
	id                 string
	WorkCount          int
	Seeds              []*spider.Task
	Fetcher            spider.Fetcher
	Storage            spider.DataRepository
	Logger             *zap.Logger
	scheduler          Scheduler
	reqRepository      spider.ReqHistoryRepository
	resourceRepository spider.ResourceRepository
	resourceRegistry   spider.ResourceRegistry
}

var defaultOptions = options{
	Logger: zap.NewNop(),
}

func WithID(id string) Option {
	return func(opts *options) {
		opts.id = id
	}
}

func WithStorage(s spider.DataRepository) Option {
	return func(opts *options) {
		opts.Storage = s
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opts *options) {
		opts.Logger = logger
	}
}
func WithFetcher(fetcher spider.Fetcher) Option {
	return func(opts *options) {
		opts.Fetcher = fetcher
	}
}

func WithWorkCount(workCount int) Option {
	return func(opts *options) {
		opts.WorkCount = workCount
	}
}

func WithSeeds(seed []*spider.Task) Option {
	return func(opts *options) {
		opts.Seeds = seed
	}
}

func WithScheduler(scheduler Scheduler) Option {
	return func(opts *options) {
		opts.scheduler = scheduler
	}
}

func WithReqRepository(reqRepository spider.ReqHistoryRepository) Option {
	return func(opts *options) {
		opts.reqRepository = reqRepository
	}
}

func WithResourceRepository(resourceRepository spider.ResourceRepository) Option {
	return func(opts *options) {
		opts.resourceRepository = resourceRepository
	}
}

func WithResourceRegistry(resourceRegistry spider.ResourceRegistry) Option {
	return func(opts *options) {
		opts.resourceRegistry = resourceRegistry
	}
}
