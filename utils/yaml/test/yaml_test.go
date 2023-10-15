package yaml

import (
	yaml "github.com/flyerxp/lib/utils/yaml"
	"os"
	"testing"
)

// Default
type App struct {
	App struct {
		Name string `yaml:"Name"`
	} `yaml:"App"`
	ZK []Zookeeper `yaml:"ZK"`
}

// ZK
type Zookeeper struct {
	Host struct {
		Addr string `yaml:"Addr"`
	} `yaml:"Host"`
}

// 普通Encode测试
func TestEncode(t *testing.T) {
	app := new(App)
	err := yaml.DecodeByFile("dev.yml", app)
	t.Logf("%v", app)
	if err != nil {
		t.Error(err.Error())
	}
	yamlString := `App:
  Name: "CH123_ADMIN"
ZK:
  - Host:
      Addr: 127.0.0.1:2181
  - Host:
      Addr: 127.0.0.1:2182`
	app2 := new(App)
	_ = yaml.DecodeByBytes([]byte(yamlString), app2)
	t.Logf("%v", app2)
	if err != nil {
		t.Error(err.Error())
	}
	yamlString, err = yaml.Encode(app2)
	t.Log(yamlString)
	if err != nil {
		t.Error(err.Error())
	}
	wf, err := os.OpenFile("tmp.yml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	_, _ = wf.WriteString(yamlString)
	_ = wf.Close()
}
