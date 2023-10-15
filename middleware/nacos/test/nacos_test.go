package nacos

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/nacos"
	yaml "github.com/flyerxp/lib/utils/yaml"
	"testing"
)

func TestConf(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	a, e := nacos.GetEngine(ctx, "nacosConf")
	r, _ := a.GetConfig(ctx, "zk", "zookeeper", "62c3bcf9-7948-4c26-a353-cebc0a7c9712")
	zk := new(config.ZookeeperConf)
	_ = yaml.DecodeByBytes(r, zk)
	fmt.Println(zk, e)
	/*a.PutService(context.Background(), nacos.ServiceRequest{
		Ip:          "127.0.0.1",
		Port:        8888,
		ClusterName: "cTest",
		ServiceName: "sTest",
		GroupName:   "gTest",
		NamespaceId: "62c3bcf9-7948-4c26-a353-cebc0a7c9712",
		Healthy:     true,
		Weight:      0,
		Metadata:    map[string]string{"call": "bbb"},
	})*/

}
