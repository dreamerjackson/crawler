package main

import (
	"github.com/dreamerjackson/crawler/collect"
	"github.com/dreamerjackson/crawler/engine"
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/log"
	"github.com/dreamerjackson/crawler/proxy"
	"github.com/dreamerjackson/crawler/storage"
	"github.com/dreamerjackson/crawler/storage/sqlstorage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"time"
)

func main() {
	// log
	plugin := log.NewStdoutPlugin(zapcore.DebugLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// set zap global logger
	zap.ReplaceGlobals(logger)

	// proxy
	proxyURLs := []string{"http://127.0.0.1:8888", "http://127.0.0.1:8888"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		logger.Error("RoundRobinProxySwitcher failed")
		return
	}

	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		Proxy:   p,
	}

	var storage storage.Storage
	storage, err = sqlstorage.New(
		sqlstorage.WithSqlUrl("root:123456@tcp(127.0.0.1:3326)/crawler?charset=utf8"),
		sqlstorage.WithLogger(logger.Named("sqlDB")),
		sqlstorage.WithBatchCount(2),
	)
	if err != nil {
		logger.Error("create sqlstorage failed")
		return
	}

	//2秒钟1个
	secondLimit := rate.NewLimiter(limiter.Per(1, 2*time.Second), 1)
	//60秒20个
	minuteLimit := rate.NewLimiter(limiter.Per(20, 1*time.Minute), 20)
	multiLimiter := limiter.MultiLimiter(secondLimit, minuteLimit)

	seeds := make([]*collect.Task, 0, 1000)
	seeds = append(seeds, &collect.Task{
		Property: collect.Property{
			Name: "douban_book_list",
		},
		Fetcher: f,
		Storage: storage,
		Limit:   multiLimiter,
	})

	s := engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)

	s.Run()

}
