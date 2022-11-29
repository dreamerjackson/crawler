package main

import (
	"context"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dreamerjackson/crawler/collect"
	"github.com/dreamerjackson/crawler/engine"
	"github.com/dreamerjackson/crawler/limiter"
	"github.com/dreamerjackson/crawler/log"
	pb "github.com/dreamerjackson/crawler/proto/greeter"
	"github.com/dreamerjackson/crawler/proxy"
	"github.com/dreamerjackson/crawler/storage"
	"github.com/dreamerjackson/crawler/storage/sqlstorage"
	"github.com/go-micro/plugins/v4/config/encoder/toml"
	etcdReg "github.com/go-micro/plugins/v4/registry/etcd"
	gs "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

func main() {

	var (
		err     error
		logger  *zap.Logger
		p       proxy.Func
		storage storage.Storage
	)

	// load config
	enc := toml.NewEncoder()
	cfg, err := config.NewConfig(config.WithReader(json.NewReader(reader.WithEncoder(enc))))
	err = cfg.Load(file.NewSource(
		file.WithPath("/Users/jackson/career/crawler/config.toml"),
		source.WithEncoder(enc),
	))

	if err != nil {
		return
	}

	// log
	logText := cfg.Get("logLevel").String("INFO")
	logLevel, err := zapcore.ParseLevel(logText)
	if err != nil {
		return
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
	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: time.Duration(timeout) * time.Millisecond,
		Logger:  logger,
		Proxy:   p,
	}

	// storage
	sqlURL := cfg.Get("storage", "sqlURL").String("")
	if storage, err = sqlstorage.New(
		sqlstorage.WithSQLURL(sqlURL),
		sqlstorage.WithLogger(logger.Named("sqlDB")),
		sqlstorage.WithBatchCount(2),
	); err != nil {
		logger.Error("create sqlstorage failed", zap.Error(err))
		return
	}

	// speed limiter
	secondLimit := rate.NewLimiter(limiter.Per(1, 2*time.Second), 1)
	minuteLimit := rate.NewLimiter(limiter.Per(20, 1*time.Minute), 20)
	multiLimiter := limiter.Multi(secondLimit, minuteLimit)

	// init tasks
	seeds := make([]*collect.Task, 0, 1000)
	seeds = append(seeds, &collect.Task{
		Property: collect.Property{
			Name: "douban_book_list",
		},
		Fetcher: f,
		Storage: storage,
		Limit:   multiLimiter,
	})

	_ = engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)

	// worker start
	//go s.Run()

	var sconfig ServerConfig
	if err := cfg.Get("GRPCServer").Scan(&sconfig); err != nil {
		logger.Error("get GRPC Server config failed", zap.Error(err))
	}
	logger.Sugar().Debugf("grpc server config,%+v", sconfig)

	// start http proxy to GRPC
	go RunHTTPServer(sconfig)

	// start grpc server
	RunGRPCServer(logger, sconfig)
}

type ServerConfig struct {
	GRPCListenAddress string
	HTTPListenAddress string
	ID                string
	RegistryAddress   string
	RegisterTTL       int
	RegisterInterval  int
	Name              string
	ClientTimeOut     int
}

func RunGRPCServer(logger *zap.Logger, cfg ServerConfig) {
	reg := etcdReg.NewRegistry(registry.Addrs(cfg.RegistryAddress))
	service := micro.NewService(
		micro.Server(gs.NewServer(
			server.Id(cfg.ID),
		)),
		micro.Address(cfg.GRPCListenAddress),
		micro.Registry(reg),
		micro.RegisterTTL(time.Duration(cfg.RegisterTTL)*time.Second),
		micro.RegisterInterval(time.Duration(cfg.RegisterInterval)*time.Second),
		micro.WrapHandler(logWrapper(logger)),
		micro.Name(cfg.Name),
	)

	// 设置micro 客户端默认超时时间为10秒钟
	if err := service.Client().Init(client.RequestTimeout(time.Duration(cfg.ClientTimeOut) * time.Second)); err != nil {
		logger.Sugar().Error("micro client init error. ", zap.String("error:", err.Error()))

		return
	}

	service.Init()

	if err := pb.RegisterGreeterHandler(service.Server(), new(Greeter)); err != nil {
		logger.Fatal("register handler failed")
	}

	if err := service.Run(); err != nil {
		logger.Fatal("grpc server stop")
	}
}

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *pb.Request, rsp *pb.Response) error {
	rsp.Greeting = "Hello " + req.Name

	return nil
}

func RunHTTPServer(cfg ServerConfig) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := pb.RegisterGreeterGwFromEndpoint(ctx, mux, cfg.GRPCListenAddress, opts); err != nil {
		zap.L().Fatal("Register backend grpc server endpoint failed")
	}
	zap.S().Debugf("start http server listening on %v proxy to grpc server;%v", cfg.HTTPListenAddress, cfg.GRPCListenAddress)
	if err := http.ListenAndServe(cfg.HTTPListenAddress, mux); err != nil {
		zap.L().Fatal("http listenAndServe failed")
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
