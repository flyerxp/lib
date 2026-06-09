package gormL

import (
	"errors"
	"gorm.io/gorm"
)

type Tx struct {
	db     *gorm.DB
	closed bool // 是否已经提交或者回滚
}

// NewTx 创建事务包装器，自动开启事务
func NewTx(db *gorm.DB) (*Tx, error) {
	txDb := db.Begin()
	if txDb.Error != nil {
		return nil, txDb.Error
	}
	return &Tx{db: txDb}, nil
}

// DB 返回底层的 *gorm.DB 事务对象
func (t *Tx) GetTxDB() *gorm.DB {
	return t.db
}

// Commit 提交事务，重复调用会返回错误
func (t *Tx) Commit() error {
	if t.closed {
		return errors.New("transaction already committed or rolled back")
	}
	if t.db == nil {
		return errors.New("gormL: db is nil")
	}

	t.closed = true
	return t.db.Commit().Error
}

// Rollback 回滚事务，重复调用会返回错误
func (t *Tx) Rollback() error {
	if t.closed {
		return errors.New("transaction already committed or rolled back")
	}
	if t.db == nil {
		return errors.New("gormL: db is nil")
	}

	t.closed = true
	return t.db.Rollback().Error
}

// Close 用于 defer，自动回滚未提交的事务
func (t *Tx) Close() error {
	if t.closed {
		return nil
	}
	if t.db == nil {
		return nil
	}

	t.closed = true
	return t.db.Rollback().Error
}

// IsClosed 检查事务是否已结束
func (t *Tx) IsClosed() bool {
	return t.closed
}
