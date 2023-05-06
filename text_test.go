package main

import (
	"log"
	"testing"
)

type Student struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func TestData(t *testing.T) {
	e := &ExcelConfig{
		context: make(chan []interface{}, 10240),
		db:      Db.Table(Test{}.TableName()).Select([]string{"userId", "age", "context"}).Limit(1000000).Offset(0),
	}
	log.Println(e.excelFileMain("./api_export.xlsx", true, []string{"userId", "age", "context"}, []string{}))

}
