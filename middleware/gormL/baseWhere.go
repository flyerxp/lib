package gormL

import "gorm.io/gorm"

// BaseWhere 通用基类
// 不依赖任何子类，全项目所有Where都能继承
type BaseWhere struct {
	wheres []func(db *gorm.DB) *gorm.DB
}

// AddWhere 内部添加条件（通用）
func (b *BaseWhere) addWhere(fn func(db *gorm.DB) *gorm.DB) {
	b.wheres = append(b.wheres, fn)
}

// Build 构建所有条件到 DB
func (b *BaseWhere) Build(db *gorm.DB) *gorm.DB {
	for _, fn := range b.wheres {
		db = fn(db)
	}
	return db
}

// Where 通用查询方法
// 子类直接调用，返回自己（实现链式）
func (b *BaseWhere) where(query string, args ...any) {
	b.addWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
}
