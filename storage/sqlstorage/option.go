package sqlstorage

import (
	"go.uber.org/zap"
)

type options struct {
	logger     *zap.Logger
	sqlURL     string
	BatchCount int // 批量数
}

var defaultOptions = options{
	logger: zap.NewNop(),
}

type Option func(opts *options)

func WithLogger(logger *zap.Logger) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func WithSQLURL(sqlURL string) Option {
	return func(opts *options) {
		opts.sqlURL = sqlURL
	}
}

func WithBatchCount(batchCount int) Option {
	return func(opts *options) {
		opts.BatchCount = batchCount
	}
}
