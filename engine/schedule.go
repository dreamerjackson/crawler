package engine

import (
	"github.com/dreamerjackson/crawler/collect"
	"go.uber.org/zap"
)

type Schedule struct {
	requestCh chan *collect.Request
	workerCh  chan *collect.Request
	out       chan collect.ParseResult
	options
}

type Config struct {
	WorkCount int
	Fetcher   collect.Fetcher
	Logger    *zap.Logger
	Seeds     []*collect.Request
}

func NewSchedule(opts ...Option) *Schedule {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	s := &Schedule{}
	s.options = options
	return s
}

func (s *Schedule) Run() {
	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)
	out := make(chan collect.ParseResult)
	s.requestCh = requestCh
	s.workerCh = workerCh
	s.out = out
	go s.Schedule()
	for i := 0; i < s.WorkCount; i++ {
		go s.CreateWork()
	}
	s.HandleResult()
}

func (s *Schedule) Schedule() {
	var reqQueue = s.Seeds
	go func() {
		for {
			var req *collect.Request
			var ch chan *collect.Request

			if len(reqQueue) > 0 {
				req = reqQueue[0]
				reqQueue = reqQueue[1:]
				ch = s.workerCh
			}
			select {
			case r := <-s.requestCh:
				reqQueue = append(reqQueue, r)

			case ch <- req:
			}
		}
	}()
}

func (s *Schedule) CreateWork() {
	for {
		r := <-s.workerCh
		if err := r.Check(); err != nil {
			s.Logger.Error("check failed",
				zap.Error(err),
			)
			continue
		}
		body, err := s.Fetcher.Get(r)
		if len(body) < 6000 {
			s.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", r.Url),
			)
			continue
		}
		if err != nil {
			s.Logger.Error("can't fetch ",
				zap.Error(err),
				zap.String("url", r.Url),
			)
			continue
		}
		result := r.ParseFunc(body, r)
		s.out <- result
	}
}

func (s *Schedule) HandleResult() {
	for {
		select {
		case result := <-s.out:
			for _, req := range result.Requesrts {
				s.requestCh <- req
			}
			for _, item := range result.Items {
				// todo: store
				s.Logger.Sugar().Info("get result: ", item)
			}
		}
	}
}
