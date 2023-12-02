package mysqlL

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Tx struct {
	db *sqlx.DB
	Tx *sqlx.Tx
}

func GetNewTx(ctx context.Context, db *sqlx.DB) *Tx {
	t := new(Tx)
	t.db = db
	return t
}
func (t *Tx) MustBegin(ctx context.Context) {
	t.Tx = t.db.MustBeginTx(ctx, nil)
}
func (t *Tx) Rollback() error {
	return t.Tx.Rollback()
}
func (t *Tx) Commit() error {
	return t.Tx.Commit()
}
func (t *Tx) Defer() {
	//异常退出回滚
	if r := recover(); r != nil {
		t.Rollback()
		panic(r)
	}
}
