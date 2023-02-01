package sqlstorage

import (
	"github.com/dreamerjackson/crawler/spider"
	"github.com/dreamerjackson/crawler/sqlstorage/sqldb"
	"github.com/dreamerjackson/crawler/tasklib/doubanbook"
	"github.com/dreamerjackson/crawler/tasklib/doubangroup"
	"github.com/dreamerjackson/crawler/tasklib/doubangroupjs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	spider.TaskStore.Add(doubangroup.DoubangroupTask)
	spider.TaskStore.Add(doubanbook.DoubanBookTask)
	spider.TaskStore.AddJSTask(doubangroupjs.DoubangroupJSTask)
}

type mysqldb struct {
}

func (m mysqldb) CreateTable(t sqldb.TableData) error {
	return nil
}

func (m mysqldb) Insert(t sqldb.TableData) error {
	return nil
}

func TestSQLStorage_Flush(t *testing.T) {
	type fields struct {
		dataDocker []*spider.DataCell
		options    options
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "empty", wantErr: false},
		{name: "no Rule filed", fields: fields{dataDocker: []*spider.DataCell{
			{Data: map[string]interface{}{"url": "http://xxx.com"}},
		}}, wantErr: true},
		{name: "no Task filed", fields: fields{dataDocker: []*spider.DataCell{
			{Data: map[string]interface{}{"url": "http://xxx.com"}},
		}}, wantErr: true},
		{name: "right data", fields: fields{dataDocker: []*spider.DataCell{
			{Data: map[string]interface{}{"Rule": "书籍简介", "Task": "douban_book_list", "Data": map[string]interface{}{"url": "http://xxx.com"}}},
		}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SQLStorage{
				dataDocker: tt.fields.dataDocker,
				db:         mysqldb{},
				options:    tt.fields.options,
			}
			if err := s.Flush(); (err != nil) != tt.wantErr {
				t.Errorf("Flush() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Nil(t, s.dataDocker)
		})
	}
}
