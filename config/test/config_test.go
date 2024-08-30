package config

import (
	"fmt"
	"github.com/flyerxp/lib/v2/config"
	"testing"
)

func TestConf(t *testing.T) {
	fmt.Println(config.GetConf().ElasticNacos)
}
