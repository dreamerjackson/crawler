package spider

type DataRepository interface {
	Save(datas ...*DataCell) error
}

type DataCell struct {
	Task *Task
	Data map[string]interface{}
}

func (d *DataCell) GetTableName() string {
	return d.Data["Task"].(string)
}

func (d *DataCell) GetTaskName() string {
	return d.Data["Task"].(string)
}
