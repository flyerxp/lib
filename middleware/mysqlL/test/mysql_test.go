package mysqlL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/mysqlL"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

type DataM struct {
	Id int `db:"id"`
}

func TestConf(T *testing.T) {
	tmpData := new(DataM)
	count := 10
	start := time.Now()
	ctx := logger.GetContext(context.Background(), "test")
	mysql, _ := mysqlL.GetEngine(ctx, "pubMysql")
	for i := 0; i <= count; i++ {
		r := rand.Intn(10000)
		sql := "select " + strconv.Itoa(r) + " as id from news_info"
		err := mysql.GetDb().GetContext(ctx, tmpData, sql)
		if err != nil {
			fmt.Println(tmpData, err)
		}
		if tmpData.Id != r {
			fmt.Println(tmpData.Id, sql)
		}
	}
	logger.WriteLine(ctx)
	fmt.Printf("mysql 数据库读取 10000次耗时 %d 毫秒\n", time.Since(start).Milliseconds())
}
func TestSelect(T *testing.T) {
	tmpData := make([]DataM, 0)
	count := 10
	start := time.Now()
	ctx := logger.GetContext(context.Background(), "test")
	mysql, _ := mysqlL.GetEngine(ctx, "pubMysql")
	for i := 0; i <= count; i++ {
		r := rand.Intn(10000)
		sql := "select " + strconv.Itoa(r) + " as id from news_info"
		err := mysql.GetDb().SelectContext(ctx, &tmpData, sql)
		fmt.Println(tmpData, err)

	}
	fmt.Printf("mysql 数据库读取 10000次耗时 %d 毫秒\n", time.Since(start).Milliseconds())
}
func TestGo(T *testing.T) {
	start := time.Now()
	ctx := logger.GetContext(context.Background(), "test")
	mysql, _ := mysqlL.GetEngine(ctx, "pubMysql")
	db := mysql.GetDb()
	wg := sync.WaitGroup{}
	for i := 0; i <= 100; i++ {
		go func() {
			wg.Add(1)
			tmpData := DataM{}
			r := strconv.Itoa(rand.Intn(10000))
			sql := "select " + r + " as id from news_info limit 1"
			err := db.GetContext(ctx, &tmpData, sql)
			if err != nil {
				fmt.Println(tmpData, err)
			} else {
				if strconv.Itoa(tmpData.Id) != r {
					fmt.Println(tmpData.Id, sql)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("mysql 数据库读取耗时 %d 毫秒\n", time.Since(start).Milliseconds())
}
func TestSql(T *testing.T) {
	T.Log(mysqlL.GetUpdateSql("abcd", []string{
		"a", "b", "c",
	}))
	T.Log(mysqlL.GetInsertSql("abcd", []string{
		"a", "b", "c",
	}))
}
func TestTx(T *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	mysql, _ := mysqlL.GetEngine(ctx, "pubMysql")
	db := mysql.GetDb()

	tx, _ := db.Begin()
	re, err := tx.ExecContext(ctx, "update news_info set title='aaaaaaaa' where id = 2")
	fmt.Println(re, "=======", err)
	tx.Rollback()
}
