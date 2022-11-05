package collect

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// 一个任务实例，
type Task struct {
	Url         string
	Cookie      string
	WaitTime    time.Duration
	MaxDepth    int
	Visited     map[string]bool
	VisitedLock sync.Mutex
	RootReq     *Request
	Fetcher     Fetcher
}

// 单个请求
type Request struct {
	unique    string
	Task      *Task
	Url       string
	Method    string
	Depth     int
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requesrts []*Request
	Items     []interface{}
}

func (r *Request) Check() error {
	if r.Depth > r.Task.MaxDepth {
		return errors.New("Max depth limit reached")
	}
	return nil
}

// 请求的唯一识别码
func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.Url + r.Method))
	return hex.EncodeToString(block[:])
}
