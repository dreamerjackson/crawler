package spider

import "fmt"

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

type EmptyDataRepository struct {
}

func (s *EmptyDataRepository) Save(dataCells ...*DataCell) error {
	for _, cell := range dataCells {
		taskName := cell.Data["Task"].(string)
		ruleName := cell.Data["Rule"].(string)
		fields := GetFields(taskName, ruleName)
		// Dynamically construct format string and collect values
		formatString := ""
		var values []interface{}
		for _, field := range fields {
			data := cell.Data["Data"].(map[string]interface{})
			value, exists := data[field]
			if !exists {
				fmt.Printf("Field %s does not exist.\n", field)
				continue
			}

			// Append to format string and values slice
			formatString += field + ": %v \n"
			values = append(values, value)
		}
		formatString += "\n"

		// Remove trailing comma and space, add newline at the end
		//if len(formatString) > 0 {
		//	formatString = formatString[:len(formatString)-2] + "\n\n"
		//}

		// Use the constructed format string and values for printing
		fmt.Printf(formatString, values...)

	}
	return nil
}
