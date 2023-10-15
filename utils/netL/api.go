package netL

import (
	"net"
)

func GetIp() []net.Addr {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []net.Addr{}
	}
	var ar []net.Addr
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ar = append(ar, address)
			}
		}
	}
	return ar
}
