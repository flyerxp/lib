package test

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var sum int32

func myFunc(i interface{}) {
	n := i.(int32)
	atomic.AddInt32(&sum, n)
	fmt.Printf("run with %d\n", n)
}

func demoFunc() {
	time.Sleep(3 * time.Second)
	fmt.Println("Hello World!")
}

func TestGo(t *testing.T) {
	var w sync.WaitGroup
	p, _ := ants.NewPool(10, ants.WithExpiryDuration(time.Second), ants.WithPreAlloc(true), ants.WithDisablePurge(true), ants.WithPanicHandler(func(a any) {
		fmt.Printf("panicHandder %#v", a)
	}))
	defer p.Release()
	for i := 1; i <= 10; i++ {
		_ = p.Submit(func() {
			w.Add(1)
			myFunc(int32(1))
			demoFunc()
			w.Done()
		})
	}

	w.Wait()
	//time.Sleep(time.Millisecond * 5000)
	//p.ReleaseTimeout(time.Second * 3)
	p.Release()
	p.Reboot()
	time.Sleep(time.Millisecond * 1)
	fmt.Println("是否已经关闭", p.IsClosed())
	fmt.Println("可用的", p.Free())
	fmt.Println("正在等待的 ", p.Waiting())
	fmt.Println("正在用的 ", p.Running())
}
