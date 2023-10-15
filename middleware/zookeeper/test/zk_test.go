package test

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/zookeeper"
	"testing"
)

func TestConf(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	zkConn := zookeeper.New(ctx, "centerZk")
	//defer zookeeper.PutConn("centerZk", zkConn)
	a, s, e := zkConn.Get("/")
	fmt.Printf("%#v", a)
	fmt.Printf("%#v", s)
	fmt.Printf("%#v", e)
	//zkConn.Close()
	zookeeper.PutConn("centerZk", zkConn)
	//time.Sleep(time.Second * 20)
	zkConn = zookeeper.New(ctx, "centerZk")
	a, s, e = zkConn.Get("/")
	fmt.Println("===================================")
	fmt.Printf("%#v", a)
	fmt.Printf("%#v", s)
	fmt.Printf("%#v", e)
	zookeeper.Reset(ctx)
}
