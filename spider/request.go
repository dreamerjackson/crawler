package spider

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
)

// 单个请求
type Request struct {
	Task     *Task
	URL      string
	Method   string
	Depth    int64
	Priority int64
	RuleName string
	TmpData  *Temp
}

func (r *Request) Check() error {
	if r.Depth > r.Task.MaxDepth {
		return errors.New("max depth limit reached")
	}

	if r.Task.Closed {
		return errors.New("task has Closed")
	}

	return nil
}

// 请求的唯一识别码
func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.URL + r.Method))

	return hex.EncodeToString(block[:])
}
