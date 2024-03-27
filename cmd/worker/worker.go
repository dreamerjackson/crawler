package worker

import (
	"context"
	"fmt"
	"github.com/dreamerjackson/crawler/generator"
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/log"
	"github.com/dreamerjackson/crawler/proto/greeter"
	"github.com/dreamerjackson/crawler/proxy"
	"github.com/dreamerjackson/crawler/spider"
	engine "github.com/dreamerjackson/crawler/spider/workerengine"
	sqlstorage "github.com/dreamerjackson/crawler/sqlstorage"
	"github.com/go-micro/plugins/v4/config/encoder/toml"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/go-micro/plugins/v4/server/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strconv"
	"time"
)

var ServiceName string = "go.micro.server.worker"

var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "run worker service.",
	Long:  "run worker service.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func init() {
	WorkerCmd.Flags().StringVar(
		&workerID, "id", "", "set worker id")

	WorkerCmd.Flags().StringVar(
		&podIP, "podip", "", "set worker id")

	WorkerCmd.Flags().StringVar(
		&HTTPListenAddress, "http", ":8080", "set HTTP listen address")

	WorkerCmd.Flags().StringVar(
		&GRPCListenAddress, "grpc", ":9090", "set GRPC listen address")

	WorkerCmd.Flags().StringVar(
		&PProfListenAddress, "pprof", ":9981", "set pprof address")

	WorkerCmd.Flags().BoolVar(
		&cluster, "cluster", true, "run mode")

}

var cluster bool

var workerID string
var HTTPListenAddress string
var GRPCListenAddress string
var PProfListenAddress string
var podIP string

func Run() {
	go func() {
		if err := http.ListenAndServe(PProfListenAddress, nil); err != nil {
			panic(err)
		}
	}()

	var (
		err     error
		logger  *zap.Logger
		p       proxy.Func
		storage spider.DataRepository
	)

	// load config
	enc := toml.NewEncoder()
	cfg, err := config.NewConfig(config.WithReader(json.NewReader(reader.WithEncoder(enc))))
	err = cfg.Load(file.NewSource(
		file.WithPath("config.toml"),
		source.WithEncoder(enc),
	))

	if err != nil {
		panic(err)
	}

	// log
	logText := cfg.Get("logLevel").String("INFO")
	logLevel, err := zapcore.ParseLevel(logText)
	if err != nil {
		panic(err)
	}
	plugin := log.NewStdoutPlugin(logLevel)
	logger = log.NewLogger(plugin)
	logger.Info("log init end")

	// set zap global logger
	zap.ReplaceGlobals(logger)

	// fetcher
	proxyURLs := cfg.Get("fetcher", "proxy").StringSlice([]string{})
	timeout := cfg.Get("fetcher", "timeout").Int(5000)
	logger.Sugar().Info("proxy list: ", proxyURLs, " timeout: ", timeout)
	if p, err = proxy.RoundRobinProxySwitcher(proxyURLs...); err != nil {
		logger.Error("RoundRobinProxySwitcher failed", zap.Error(err))
	}
	f := spider.NewFetchService(spider.BrowserFetchType)

	// storage
	storeType := cfg.Get("storage", "type").String("")
	switch storeType {
	case "mysql":
		sqlURL := cfg.Get("storage", "sqlURL").String("")
		if storage, err = sqlstorage.New(
			sqlstorage.WithSQLURL(sqlURL),
			sqlstorage.WithLogger(logger.Named("sqlDB")),
			sqlstorage.WithBatchCount(2),
		); err != nil {
			logger.Error("create sqlstorage failed", zap.Error(err))
			panic(err)
			return
		}
		logger.Info("start mysql storage")
	case "empty":
		storage = &spider.EmptyDataRepository{}
		logger.Info("start empty storage")
	}

	// init tasks
	var tcfg []spider.TaskConfig
	if err := cfg.Get("Tasks").Scan(&tcfg); err != nil {
		logger.Error("init seed tasks", zap.Error(err))
	}
	seeds := ParseTaskConfig(logger, p, f, storage, tcfg)

	var sconfig ServerConfig
	if err := cfg.Get("GRPCServer").Scan(&sconfig); err != nil {
		logger.Error("get GRPC Server config failed", zap.Error(err))
	}
	logger.Sugar().Debugf("grpc server config,%+v", sconfig)
	if workerID == "" {
		if podIP != "" {
			ip := generator.GetIDbyIP(podIP)
			workerID = strconv.Itoa(int(ip))
		} else {
			workerID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
	}

	id := sconfig.Name + "-" + workerID
	zap.S().Debug("worker id:", id)

	// init etcd registry
	reg, err := spider.NewEtcdRegistry([]string{sconfig.RegistryAddress})
	if err != nil {
		logger.Error("init EtcdRegistry failed", zap.Error(err))
	}

	s, err := engine.NewWorkerService(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
		engine.WithStorage(storage),
		engine.WithID(id),
		engine.WithReqRepository(spider.NewReqHistoryRepository()),
		engine.WithResourceRepository(spider.NewResourceRepository()),
		engine.WithResourceRegistry(reg),
	)

	if err != nil {
		panic(err)
	}

	// worker start
	go s.Run(cluster)

	// start http proxy to GRPC
	go RunHTTPServer(sconfig)

	// start grpc server
	RunGRPCServer(logger, sconfig)
}

type ServerConfig struct {
	RegistryAddress  string
	RegistryType     string
	RegisterTTL      int
	RegisterInterval int
	Name             string
	ClientTimeOut    int
}

func RunGRPCServer(logger *zap.Logger, cfg ServerConfig) {

	var options = []micro.Option{
		micro.Server(grpc.NewServer(
			server.Id(workerID),
		)),
		micro.Address(GRPCListenAddress),
		micro.RegisterTTL(time.Duration(cfg.RegisterTTL) * time.Second),
		micro.RegisterInterval(time.Duration(cfg.RegisterInterval) * time.Second),
		micro.WrapHandler(logWrapper(logger)),
		micro.Name(cfg.Name),
	}

	switch cfg.RegistryType {
	case "etcd":
		reg := etcd.NewRegistry(registry.Addrs(cfg.RegistryAddress))
		options = append(options, micro.Registry(reg))
	}

	service := micro.NewService(
		options...,
	)

	// 设置micro 客户端默认超时时间为10秒钟
	if err := service.Client().Init(client.RequestTimeout(time.Duration(cfg.ClientTimeOut) * time.Second)); err != nil {
		logger.Sugar().Error("micro client init error. ", zap.String("error:", err.Error()))

		return
	}

	service.Init()

	if err := greeter.RegisterGreeterHandler(service.Server(), new(Greeter)); err != nil {
		logger.Fatal("register handler failed", zap.Error(err))
	}

	if err := service.Run(); err != nil {
		logger.Fatal("grpc server stop", zap.Error(err))
	}
}

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *greeter.Request, rsp *greeter.Response) error {
	rsp.Greeting = "Hello " + req.Name

	return nil
}

func RunHTTPServer(cfg ServerConfig) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc2.DialOption{
		grpc2.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := greeter.RegisterGreeterGwFromEndpoint(ctx, mux, GRPCListenAddress, opts); err != nil {
		zap.L().Fatal("Register backend grpc server endpoint failed", zap.Error(err))
	}
	zap.S().Debugf("start http server listening on %v proxy to grpc server;%v", HTTPListenAddress, GRPCListenAddress)
	if err := http.ListenAndServe(HTTPListenAddress, mux); err != nil {
		zap.L().Fatal("http listenAndServe failed", zap.Error(err))
	}
}

func logWrapper(log *zap.Logger) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {

			log.Info("receive request",
				zap.String("method", req.Method()),
				zap.String("Service", req.Service()),
				zap.Reflect("request param:", req.Body()),
			)

			err := fn(ctx, req, rsp)

			return err
		}
	}
}

func ParseTaskConfig(logger *zap.Logger, p proxy.Func, f spider.Fetcher, s spider.DataRepository, cfgs []spider.TaskConfig) []*spider.Task {
	tasks := make([]*spider.Task, 0, 1000)
	for _, cfg := range cfgs {
		t := spider.NewTask(
			spider.WithName(cfg.Name),
			spider.WithReload(cfg.Reload),
			spider.WithCookie(cfg.Cookie),
			spider.WithLogger(logger),
			spider.WithStorage(s),
			spider.WithProxy(p),
		)

		if cfg.WaitTime > 0 {
			t.WaitTime = cfg.WaitTime
		}

		if cfg.MaxDepth > 0 {
			t.MaxDepth = cfg.MaxDepth
		}

		var limits []limiter.RateLimiter
		if len(cfg.Limits) > 0 {
			for _, lcfg := range cfg.Limits {
				// speed limiter
				l := rate.NewLimiter(limiter.Per(lcfg.EventCount, time.Duration(lcfg.EventDur)*time.Second), lcfg.Bucket)
				limits = append(limits, l)
			}
			multiLimiter := limiter.Multi(limits...)
			t.Limit = multiLimiter
		}

		t.Fetcher = f
		tasks = append(tasks, t)
	}
	return tasks
}
