package netL

import (
	"fmt"
	"github.com/flyerxp/lib/utils/netL"
	"net"
	"testing"
)

func TestIp(T *testing.T) {
	add := netL.GetIp()
	fmt.Println(add[0].(*net.IPNet).IP)
}
