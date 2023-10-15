package app

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/app"
	"strconv"
	"testing"
)

func TestFun(t *testing.T) {
	ctx := context.Background()
	defer func() { app.Shutdown(ctx) }()
	for i := 0; i <= 20; i++ {
		app.RegisterFunc("注冊"+strconv.Itoa(i), "注冊描述"+strconv.Itoa(i), func() {
			fmt.Println("我執行了" + strconv.Itoa(i))
		})
	}
}
