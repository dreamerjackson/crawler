package sqldb

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"strings"
)

type DBer interface {
	CreateTable(t TableData) error
	Insert(t TableData) error
}

type Sqldb struct {
	options
	db *sql.DB
}

type Field struct {
	Title string
	Type  string
}

type TableData struct {
	TableName   string
	ColumnNames []Field       // 标题字段
	Args        []interface{} // 数据
	DataCount   int           // 插入数据的数量
	AutoKey     bool
}

func New(opts ...Option) (*Sqldb, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	d := &Sqldb{}
	d.options = options

	if err := d.OpenDB(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Sqldb) OpenDB() error {
	db, err := sql.Open("mysql", d.sqlURL)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(2048)
	db.SetMaxIdleConns(2048)

	if err = db.Ping(); err != nil {
		return err
	}

	d.db = db

	return nil
}

func (d *Sqldb) CreateTable(t TableData) error {
	if len(t.ColumnNames) == 0 {
		return errors.New("column can not be empty")
	}

	sql := `CREATE TABLE IF NOT EXISTS ` + t.TableName + " ("

	if t.AutoKey {
		sql += `id INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT,`
	}

	for _, t := range t.ColumnNames {
		sql += t.Title + ` ` + t.Type + `,`
	}

	sql = sql[:len(sql)-1] + `) ENGINE=MyISAM DEFAULT CHARSET=utf8;`

	d.logger.Debug("crate table", zap.String("sql", sql))

	_, err := d.db.Exec(sql)

	return err
}

func (d *Sqldb) DropTable(t TableData) error {
	if len(t.ColumnNames) == 0 {
		return errors.New("column can not be empty")
	}

	sql := `DROP TABLE ` + t.TableName

	d.logger.Debug("drop table", zap.String("sql", sql))

	_, err := d.db.Exec(sql)

	return err
}

func (d *Sqldb) Insert(t TableData) error {
	if len(t.ColumnNames) == 0 {
		return errors.New("empty column")
	}

	sql := `INSERT INTO ` + t.TableName + `(`

	for _, v := range t.ColumnNames {
		sql += v.Title + ","
	}

	sql = sql[:len(sql)-1] + `) VALUES `

	blank := ",(" + strings.Repeat(",?", len(t.ColumnNames))[1:] + ")"
	sql += strings.Repeat(blank, t.DataCount)[1:] + `;`
	d.logger.Debug("insert table", zap.String("sql", sql))
	_, err := d.db.Exec(sql, t.Args...)

	return err
}
