package main

import (
	"github.com/sourcegraph/conc/iter"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"sort"
	"strconv"
	"sync"
)

type Test struct {
	gorm.Model
	UserId  string `json:"userId"`
	Age     int    `json:"age"`
	Context string `json:"context"`
}

func (Test) TableName() string {
	return "test"
}

var Db *gorm.DB

func init() {
	dsn := "root:1234@tcp(127.0.0.1:3306)/gva?charset=utf8mb4&parseTime=True&loc=Local"
	Db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	err := Db.AutoMigrate(&Test{})
	if err != nil {
		log.Println(err)
	}
}

type ExcelConfig struct {
	once        sync.Once
	db          *gorm.DB
	excelWriter *excelize.StreamWriter
	wg          sync.WaitGroup
	look        sync.Mutex
	setLock     sync.Mutex
	size        int
	context     chan []interface{}
}

func (t *ExcelConfig) head(keys []string) (err error) {
	//defer t.wg.Done()
	t.once.Do(func() {
		sS := make([]interface{}, len(keys))
		for i := 0; i < len(keys); i++ {
			sS[i] = keys[i]
		}
		err = t.excelWriter.SetRow("A1", sS)
	})
	return
}
func (t *ExcelConfig) excelFileMain(path string, allData bool, fields, head []string) (err error) {
	//t.wg = sync.WaitGroup{}
	//t.wg.Add(1)
	m := make([]map[string]interface{}, t.db.RowsAffected)
	t.db.Find(&m)
	file := excelize.NewFile()
	defer file.Close()
	t.excelWriter, err = file.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	if allData {
		go func() {
			i := 0
			for v := range t.context {
				t.excelWriter.SetRow("A"+strconv.Itoa(i+2), v)
				i++
				log.Println(i)
				if i == 100 {
					return
				}
			}
		}()
		log.Println("进行到这一步了")
		iter.ForEach(m, func(v *map[string]interface{}) {
			keys := make([]string, 0)
			for k1, _ := range *v {
				keys = append(keys, k1)
			}
			myMap := make([]interface{}, 0)
			sort.Strings(keys)
			err = t.head(keys)
			x := *v
			for _, k1 := range keys {
				myMap = append(myMap, x[k1])
			}
			t.context <- myMap
			//t.look.Lock()
			//k := t.size
			//t.size++
			//t.look.Unlock()
			//t.excelWriter.SetRow("A"+strconv.Itoa(k+2), myMap)

		})
		//for k, v := range m {
		//	keys := make([]string, 0)
		//	for k1, _ := range v {
		//		keys = append(keys, k1)
		//	}
		//	myMap := make([]interface{}, 0)
		//	sort.Strings(keys)
		//	err = t.head(keys)
		//	for _, k1 := range keys {
		//		myMap = append(myMap, v[k1])
		//	}
		//	t.excelWriter.SetRow("A"+strconv.Itoa(k+2), myMap)
		//}

	}
	if allData != true {
		for k, v := range m {
			row := make([]interface{}, len(fields))

			for i := 0; i < len(v); i++ {
				row[i] = v[fields[i]]
			}
			t.excelWriter.SetRow("A"+strconv.Itoa(k+2), row)
		}
	}

	err = t.excelWriter.Flush()
	if err != nil {
		return err
	}
	err = file.SaveAs(path)
	//t.wg.Wait()
	if err != nil {
		return err
	}
	return
}
