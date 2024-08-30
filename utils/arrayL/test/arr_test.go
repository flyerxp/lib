package test

import (
	"fmt"
	"github.com/flyerxp/lib/v2/utils/arrayL"
	"testing"
)

func TestArrConv(t *testing.T) {
	ids32 := []int32{1, 2, 3, 1, 2, 3}
	ids := []int{1, 2, 3, 1, 2, 3}
	fmt.Printf("%#v\n", arrayL.UniqArr(ids))
	fmt.Printf("%#v\n", arrayL.UniqArr32(ids32))
	fmt.Printf("%#v\n", arrayL.UniqArr32ToInt(ids32))
	fmt.Printf("%#v\n", arrayL.UniqArrToString(ids))
	fmt.Printf("%#v\n", arrayL.UniqArr32ToInt(ids32))
	fmt.Printf("%#v\n", arrayL.UniqArr32ToString(ids32))
}
