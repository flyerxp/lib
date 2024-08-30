package config

import (
	"github.com/flyerxp/lib/v2/utils/env"
	"github.com/flyerxp/lib/v2/utils/json"
	yaml "github.com/flyerxp/lib/v2/utils/yaml"
	_ "go.uber.org/zap/zapcore"
	"log"
	"path/filepath"
	"sync"
)

var (
	prefix = "conf"
	conf   *Config
	once   sync.Once
)

type zapConfig struct {
	Level            string            `yaml:"level" json:"level"`
	Encoding         string            `yaml:"encoding" json:"encoding"`
	OutputPaths      []string          `yaml:"outputPaths" json:"outputPaths"`
	ErrorOutputPaths []string          `yaml:"errorOutputPaths" json:"errorOutputPaths"`
	InitialFields    map[string]string `yaml:"initialFields" json:"initialFields"`
	EncoderConfig    map[string]string `yaml:"encoderConfig" json:"encoderConfig"`
}
type Config struct {
	Env string `yaml:"env" json:"env"`
	App struct {
		Name        string    `yaml:"name" json:"name"`
		Type        string    `yaml:"type" json:"type"`
		Logger      zapConfig `yaml:"logger" json:"logger"`
		ErrLog      zapConfig `yaml:"errlog" json:"errlog"`
		ConfStorage bool      `yaml:"confStorage" json:"confStorage"`
	}
	Hertz          Hertz           `yaml:"hertz" json:"hertz"`
	Redis          []MidRedisConf  `yaml:"redis" json:"redis"`
	RedisNacos     NacosConf       `yaml:"redisNacos" json:"redisNacos"`
	Mysql          []MidMysqlConf  `yaml:"mysql" json:"mysql"`
	MysqlNacos     NacosConf       `yaml:"mysqlNacos" json:"mysqlNacos"`
	Pulsar         []MidPulsarConf `yaml:"pulsar" json:"pulsar"`
	PulsarNacos    NacosConf       `yaml:"pulsarNacos" json:"pulsarNacos"`
	Mqtt           []MidMqttConf   `yaml:"mqtt" json:"mqtt"`
	MqttNacos      NacosConf       `yaml:"mqttNacos" json:"mqttNacos"`
	Nacos          []MidNacos      `yaml:"nacos" json:"nacos"`
	TopicNacos     []NacosConf     `yaml:"topicNacos"`
	MqttTopicNacos []NacosConf     `yaml:"MqttTopicNacos" json:"mqttTopicNacos"`
	Elastic        []MidEsConf     `yaml:"elastic" json:"elastic"`
	ElasticNacos   NacosConf       `yaml:"elasticNacos" json:"elasticNacos"`
}

func GetLogIdKey() string {
	return "loggerId"
}
func (c *Config) String() string {
	b, e := json.Encode(c)
	if e != nil {
		log.Fatalf("config josn error %s", e)
	}
	return string(b)
}

func init() {
	go GetConf()
}
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

// func (a *Config) getLoggerConf() zap.Config {
// return a.App.Logger
// }
/*func (a *Config) getRedisConf(name string) {
	if c.Redis == nil {
		err := yaml.DecodeByFile(filepath.Join(prefix, filepath.Join(env.GetEnv(), "redis.yml")), config)
		if err != nil {
			//logger.Logger.
		}
	}
}*/

var defaultConfig = []byte(`
env: test
app:
  name: Webhook
  type: web
  logger:
    level: debug
    encoding: json
    outputPaths:
      - stdout
      - logs/webhook
    errorOutputPaths:
      - stderr
    initialFields:
      app: Webhook
    encoderConfig:
      #messageKey: msg
      levelKey: level
      nameKey: name
      TimeKey: time
      #CallerKey: caller
      #FunctionKey: func
      StacktraceKey: stacktrace
      LineEnding: "\n"
  errlog:
    level: warn
    encoding: json
    outputPaths:
      - stdout
      #- logs/webhook
    errorOutputPaths:
      - stderr
    initialFields:
      app: Webhook
    encoderConfig:
      #messageKey: msg
      levelKey: level
      nameKey: name
      TimeKey: time
      CallerKey: caller
      FunctionKey: func
      StacktraceKey: stacktrace
      LineEnding: "\n"
redis:
-
  name: pubRedis
  address: [ "pubredis:6379" ]
  user:
  pwd: 123456
  master:
redisNacos:
  name: nacosConf
  did: redis
  group: redis
  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
mysql:
-
  name: pubMysql
  address: pubmysql
  port: 3306
  user: test
  pwd: 123456
  ssl: disable
  db: ch123
  sql_log: yes
  read_timeout: 100
  conn_timeout: 100
  write_timeout: 100
  char_set: utf8mb4
  max_idle_conns: 10
  max_open_conns: 500
mysqlNacos:
  name: nacosConf
  did: mysql
  group: mysql
  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
pulsar:
-
  name: pubPulsar
  address: [ "pubpulsar:6650" ]
  user: admin
  pwd: 123456
  token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJtcXR0In0.C_Uy7A-eDL43sJPyDqV1LXOXKTZI2djECeO93o6JQBE  
pulsarNacos:
  name: nacosConf
  did: pulsar
  group: pulsar
  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
mqtt:
-
  name: pubMqtt
  address: [ "pubmqtt:1883" ]
  user: 
  pwd: 
#mqttNacos:
#  name: nacosConf
#  did: mqtt
#  group: mqtt
#  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
topicNacos:
-
  name: nacosConf
  did: topic
  group: topic
  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
nacos:
-
  name: nacosConf
  url: http://nacosconf:8848/nacos
  contextPath: /nacos
  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
  user: dev
  pwd: 123456
  redis:
    name: base
    address: [ "pubEs:6379" ]
    user: 123456
    pwd: 
    master:
elastic:
-
  name: pubEs
  host: [ "pubEs:9200" ]
  user: elastic
  pwd: SLmOE+pJcwsxbFrf-rzh
  read_timeout: 600ms
  conn_timeout: 80ms
  default_max_window_result: 0
  default_track_total_hits: 0
  auto_detect: false
  max_window_result:
    test: 1000
  track_total_hits:
    test: 2000
#elasticNacos:
#  name: nacosConf
#  did: elastic
#  group: elastic
#  ns: 62c3bcf9-7948-4c26-a353-cebc0a7c9712
`)

func initConf() {
	conf = new(Config)
	err := yaml.DecodeByFile(env.GetConfRoot()+filepath.Join(prefix, filepath.Join(env.GetEnv(), "app.yml")), conf)
	if err != nil {
		log.Printf("yaml err %v", err)
		log.Print("use default config")
		err = yaml.DecodeByBytes(defaultConfig, conf)
		if err != nil {
			log.Printf("default config error", err)
		}
	}

}
func GetConfFile(s string) string {
	return env.GetConfRoot() + filepath.Join(prefix, filepath.Join(env.GetEnv(), s))
}
