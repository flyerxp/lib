package config

import (
	"fmt"
	"github.com/flyerxp/lib/config"
	"testing"
)

func TestConf(t *testing.T) {
	fmt.Println(config.GetConf().ElasticNacos)
}
