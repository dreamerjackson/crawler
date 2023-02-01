package sqlstorage

import (
	"encoding/json"
	"errors"
	"github.com/dreamerjackson/crawler/spider"
	"github.com/dreamerjackson/crawler/sqlstorage/sqldb"

	"go.uber.org/zap"
)

type SQLStorage struct {
	dataDocker []*spider.DataCell // 分批输出结果缓存
	db         sqldb.DBer
	Table      map[string]struct{}
	options
}

func New(opts ...Option) (*SQLStorage, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	s := &SQLStorage{}
	s.options = options
	s.Table = make(map[string]struct{})

	var err error
	s.db, err = sqldb.New(
		sqldb.WithConnURL(s.sqlURL),
		sqldb.WithLogger(s.logger),
	)

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SQLStorage) Save(dataCells ...*spider.DataCell) error {
	for _, cell := range dataCells {
		name := cell.GetTableName()
		if _, ok := s.Table[name]; !ok {
			// 创建表
			columnNames := getFields(cell)

			err := s.db.CreateTable(sqldb.TableData{
				TableName:   name,
				ColumnNames: columnNames,
				AutoKey:     true,
			})
			if err != nil {
				s.logger.Error("create table falied", zap.Error(err))
				continue
			}

			s.Table[name] = struct{}{}
		}

		if len(s.dataDocker) >= s.BatchCount {
			if err := s.Flush(); err != nil {
				s.logger.Error("insert data failed", zap.Error(err))
			}
		}

		s.dataDocker = append(s.dataDocker, cell)
	}

	return nil
}

func getFields(cell *spider.DataCell) []sqldb.Field {
	taskName := cell.Data["Task"].(string)
	ruleName := cell.Data["Rule"].(string)
	fields := spider.GetFields(taskName, ruleName)

	var columnNames []sqldb.Field
	for _, field := range fields {
		columnNames = append(columnNames, sqldb.Field{
			Title: field,
			Type:  "MEDIUMTEXT",
		})
	}

	columnNames = append(columnNames,
		sqldb.Field{Title: "URL", Type: "VARCHAR(255)"},
		sqldb.Field{Title: "Time", Type: "VARCHAR(255)"},
	)

	return columnNames
}

func (s *SQLStorage) Flush() error {
	if len(s.dataDocker) == 0 {
		return nil
	}

	defer func() {
		s.dataDocker = nil
	}()

	args := make([]interface{}, 0)
	var ruleName string
	var taskName string
	var ok bool
	for _, datacell := range s.dataDocker {
		if ruleName, ok = datacell.Data["Rule"].(string); !ok {
			return errors.New("no rule field")
		}

		if taskName, ok = datacell.Data["Task"].(string); !ok {
			return errors.New("no task field")
		}
		fields := spider.GetFields(taskName, ruleName)

		data := datacell.Data["Data"].(map[string]interface{})
		value := []string{}

		for _, field := range fields {
			v := data[field]
			switch v := v.(type) {
			case nil:
				value = append(value, "")
			case string:
				value = append(value, v)
			default:
				j, err := json.Marshal(v)
				if err != nil {
					value = append(value, "")
				} else {
					value = append(value, string(j))
				}
			}
		}
		if v, ok := datacell.Data["URL"].(string); ok {
			value = append(value, v)
		}
		if v, ok := datacell.Data["Time"].(string); ok {
			value = append(value, v)
		}
		for _, v := range value {
			args = append(args, v)
		}
	}

	return s.db.Insert(sqldb.TableData{
		TableName:   s.dataDocker[0].GetTableName(),
		ColumnNames: getFields(s.dataDocker[0]),
		Args:        args,
		DataCount:   len(s.dataDocker),
	})
}
