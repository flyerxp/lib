package gormL

import "gorm.io/gorm"

// BaseWhere 通用基类
// 不依赖任何子类，全项目所有Where都能继承
type BaseWhere struct {
	Wheres []func(db *gorm.DB) *gorm.DB
}

func GetBaseWhere() *BaseWhere {
	return &BaseWhere{}
}

// AddWhere 内部添加条件（通用）
func (b *BaseWhere) AddWhere(fn func(db *gorm.DB) *gorm.DB) {
	b.Wheres = append(b.Wheres, fn)
}

// Build 构建所有条件到 DB
func (b *BaseWhere) Build(db *gorm.DB) *gorm.DB {
	for _, fn := range b.Wheres {
		db = fn(db)
	}
	return db
}

// Where 通用查询方法
// 子类直接调用，返回自己（实现链式）
func (b *BaseWhere) Where(query string, args ...any) {
	b.AddWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
}
