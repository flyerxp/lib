package app

import (
	"context"
	"github.com/flyerxp/lib/v2/logger"
)

type event struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
	T      int8   `json:"t"`
	F      func()
	CtxF   func(ctx context.Context)
}

var vEvent []event

func RegisterFunc(name string, desc string, f ...func()) error {
	for i := 0; i < len(f); i++ {
		vEvent = append(vEvent, event{name, desc, true, 1, f[i], nil})
	}
	return nil
}
func RegisterCtxFunc(name string, desc string, f ...func(ctx context.Context)) error {
	for i := 0; i < len(f); i++ {
		vEvent = append(vEvent, event{name, desc, true, 2, nil, f[i]})
	}
	return nil
}
func Shutdown(ctx context.Context) {
	for i := 0; i < len(vEvent); i++ {
		if vEvent[i].Status {
			if vEvent[i].T == 1 {
				vEvent[i].F()
			} else {
				vEvent[i].CtxF(ctx)
			}
		}
	}
	logger.Sync()
}
