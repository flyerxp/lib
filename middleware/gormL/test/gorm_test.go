package gormL

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/middleware/gormL"
	"testing"
	"time"
)

// NewsInfo 表结构，与 mysqlL/test/mysql_test.go 中引用的 news_info 表对应
// 该表至少包含 id（int 主键）和 title（varchar）字段
type NewsInfo struct {
	Id    int    `gorm:"column:id;primaryKey;autoIncrement"`
	Title string `gorm:"column:title;size:255"`
}

func (NewsInfo) TableName() string {
	return "news_info"
}

// TestCreate 插入测试
func TestCreate(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	news := NewsInfo{
		Title: fmt.Sprintf("测试标题_%d", time.Now().UnixNano()),
	}
	result := db.WithContext(ctx).Create(&news)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Logf("插入成功 id=%d title=%s", news.Id, news.Title)
}

// TestFirst 查询单条记录
func TestFirst(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	var news NewsInfo
	result := db.WithContext(ctx).First(&news)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Logf("查询结果 id=%d title=%s", news.Id, news.Title)
}

// TestFind 查询多条记录
func TestFind(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	var list []NewsInfo
	result := db.WithContext(ctx).Where("id < ?", 10).Find(&list)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Logf("查询到 %d 条记录", len(list))
	for _, v := range list {
		t.Logf("  id=%d title=%s", v.Id, v.Title)
	}
}

// TestUpdate 更新测试
func TestUpdate(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	// 先查询一条记录用于更新
	var news NewsInfo
	if err := db.WithContext(ctx).First(&news).Error; err != nil {
		t.Fatal(err)
	}
	newTitle := fmt.Sprintf("updated_title_%d", time.Now().UnixNano())
	result := db.WithContext(ctx).Model(&news).Update("title", newTitle)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Logf("更新成功 id=%d 新标题=%s 影响行数=%d", news.Id, newTitle, result.RowsAffected)
}

// TestDelete 删除测试（先插入再删除，避免影响已有数据）
func TestDelete(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	// 先插入一条数据用于删除
	news := NewsInfo{
		Title: fmt.Sprintf("待删除_%d", time.Now().UnixNano()),
	}
	if err := db.WithContext(ctx).Create(&news).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("已插入待删除数据 id=%d", news.Id)
	// 执行删除
	result := db.WithContext(ctx).Delete(&news)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.RowsAffected == 0 {
		t.Fatal("删除失败，影响行数为 0")
	}
	t.Logf("删除成功 id=%d 影响行数=%d", news.Id, result.RowsAffected)
}

// TestTx 事务测试
func TestTx(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	// 开启事务
	tx := db.WithContext(ctx).Begin()
	news1 := NewsInfo{Title: fmt.Sprintf("事务插入1_%d", time.Now().UnixNano())}
	if err := tx.Create(&news1).Error; err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	news2 := NewsInfo{Title: fmt.Sprintf("事务插入2_%d", time.Now().UnixNano())}
	if err := tx.Create(&news2).Error; err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("事务提交成功 news1.id=%d news2.id=%d", news1.Id, news2.Id)
}

// TestTxRollback 事务回滚测试
func TestTxRollback(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	db, err := gormL.GetEngine(ctx, "pubMysql")
	if err != nil {
		t.Fatal(err)
	}
	// 开启事务
	tx := db.WithContext(ctx).Begin()
	news := NewsInfo{Title: fmt.Sprintf("回滚测试_%d", time.Now().UnixNano())}
	if err := tx.Create(&news).Error; err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	// 模拟错误，主动回滚
	tx.Rollback()
	t.Logf("事务已回滚，id=%d 的数据不会持久化", news.Id)

	// 确认回滚后查不到该数据
	var check NewsInfo
	result := db.WithContext(ctx).Where("id = ?", news.Id).First(&check)
	if result.Error != nil {
		t.Logf("回滚验证通过，数据不存在: %v", result.Error)
	} else {
		t.Fatal("回滚失败，数据依然存在")
	}
}
