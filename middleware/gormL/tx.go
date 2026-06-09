package gormL

import (
	"errors"
	"gorm.io/gorm"
)

type Tx struct {
	db     *gorm.DB
	closed bool
}

// Commit 提交事务，重复调用会返回错误
func (t *Tx) Commit() error {
	if t.closed {
		return errors.New("transaction already committed or rolled back")
	}

	t.closed = true
	return t.db.Commit().Error
}

// Rollback 回滚事务，重复调用会返回错误
func (t *Tx) Rollback() error {
	if t.closed {
		return errors.New("transaction already committed or rolled back")
	}
	t.closed = true
	return t.db.Rollback().Error
}

// Close 用于 defer，自动回滚未提交的事务
func (t *Tx) Close() error {
	if t.closed {
		return nil
	}
	// 未提交的事务自动回滚
	t.closed = true
	return t.db.Rollback().Error
}
