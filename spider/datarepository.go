package spider

import (
	"fmt"
	"sync"
)

type DataRepository interface {
	Save(datas ...*DataCell) error
	GetAllNews() map[string][]Item
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

type EmptyDataRepository struct{}

func (s *EmptyDataRepository) GetAllNews() map[string][]Item {
	//TODO implement me
	panic("implement me")
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

type UIRepository struct {
	Data sync.Map // 使用sync.Map存储每个任务的新闻列表
}
type Item struct {
	Fields map[string]interface{}
}

func (repo *UIRepository) SaveTo(taskName string, newsItem Item) {
	// 尝试获取现有的新闻列表
	value, ok := repo.Data.Load(taskName)
	var newsList []Item
	if ok {
		newsList = value.([]Item)
	}

	// 检查新闻项是否已存在
	exists := false
	for _, item := range newsList {
		if item.Fields["标题"] == newsItem.Fields["标题"] {
			exists = true
			break
		}
	}

	if exists {
		fmt.Printf("Duplicate found, not adding: %v\n", newsItem.Fields["标题"])
	} else {
		newsList = append(newsList, newsItem)
		repo.Data.Store(taskName, newsList)
		fmt.Printf("Added news item to %s: %v\n", taskName, newsItem.Fields["标题"])
	}

	//if !exists {
	//	newsList = append(newsList, newsItem)
	//	repo.Data.Store(taskName, newsList)
	//}
	// todo : 不能无限的append
}

func (repo *UIRepository) Save(dataCells ...*DataCell) error {

	for _, cell := range dataCells {
		taskName, ok := cell.Data["Task"].(string)
		if !ok {
			fmt.Println("Task field is missing or not a string")
			continue
		}
		//ruleName, ok := cell.Data["Rule"].(string)
		if !ok {
			fmt.Println("Rule field is missing or not a string")
			continue
		}
		//fields := GetFields(taskName, ruleName)

		items, ok := cell.Data["Data"].(map[string]interface{})
		if !ok {
			fmt.Println("Data field is missing or not a map[string]interface{}")
			continue
		}

		newsItem := Item{Fields: items}
		//for _, field := range fields {
		//	Data := cell.Data["Data"].(map[string]interface{})
		//	_, exists := Data[field]
		//	if !exists {
		//		fmt.Printf("Field %s does not exist.\n", field)
		//		continue
		//	}
		//	for field, value := range Data {
		//		newsItem.Fields[field] = value
		//	}
		//}
		// 保存新闻项目到repository
		repo.SaveTo(taskName, newsItem)

	}
	return nil
}

func (repo *UIRepository) GetAllNews() map[string][]Item {
	allNews := make(map[string][]Item)
	repo.Data.Range(func(key, value interface{}) bool {
		allNews[key.(string)] = value.([]Item)
		return true // 继续迭代
	})
	return allNews
}
