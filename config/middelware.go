package config

type MidConf struct {
	Name    string   `yaml:"name" json:"name"`
	Address []string `yaml:"address" json:"address"`
	User    string   `yaml:"user" json:"user"`
	Pwd     string   `yaml:"pwd" json:"pwd"`
}
type MidMysqlConf struct {
	Name    string `yaml:"name" json:"name"`
	Address string `yaml:"address" json:"address"`
	Port    string `yaml:"port" json:"port"`
	User    string `yaml:"user" json:"user"`
	Pwd     string `yaml:"pwd" json:"pwd"`
	Db      string `yaml:"db" json:"db"`
	//Ssl          string `yaml:"ssl" json:"ssl"` //true|false
	CharSet      string `yaml:"char_set" json:"char_set"`
	ReadTimeout  int    `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout" json:"write_timeout"`
	ConnTimeout  int    `yaml:"conn_timeout" json:"conn_timeout"`
	Collation    string `yaml:"collation" json:"collation"`
	SqlLog       string `yaml:"sql_log" json:"sql_log"` // yes|no
	MaxOpenConns int    `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns" json:"max_idle_conns"`
}
type MidPulsarConf struct {
	Name    string   `yaml:"name" json:"name"`
	Address []string `yaml:"address" json:"address"`
	User    string   `yaml:"user" json:"user"`
	Pwd     string   `yaml:"pwd" json:"pwd"`
	Token   string   `yaml:"token" json:"token"` //jwt 验证方式
}
type MysqlConf struct {
	List []MidMysqlConf `yaml:"mysql" json:"mysql"`
}
type ZookeeperConf struct {
	List []MidConf `yaml:"zookeeper" json:"zookeeper"`
}
type NacosConf struct {
	Name  string `yaml:"name" json:"name"`
	Did   string `yaml:"did" json:"did"`
	Group string `yaml:"group" json:"group"`
	Ns    string `yaml:"ns" json:"ns"`
}
type MidNacos struct {
	Name  string       `yaml:"name" json:"name"`
	Url   string       `yaml:"url" json:"url"`
	Ns    string       `yaml:"ns" json:"ns"`
	Redis MidRedisConf `yaml:"redis" json:"redis"`
	User  string       `yaml:"user" json:"user"`
	Pwd   string       `yaml:"pwd" json:"pwd"`
}
type Nacos struct {
	List []MidNacos `yaml:"nacos"`
}

type MidRedisConf struct {
	Name    string   `yaml:"name" json:"name"`
	Address []string `yaml:"address" json:"address"`
	User    string   `yaml:"user" json:"user"`
	Pwd     string   `yaml:"pwd" json:"pwd"`
	Master  string   `yaml:"master" json:"master"` //哨兵模式使用，写masterName
}
type RedisConf struct {
	Redis []MidRedisConf `yaml:"redis" json:"redis"`
}

type PulsarConf struct {
	List []MidPulsarConf `yaml:"pulsar" json:"pulsar"`
}

type MqttConf struct {
	List []MidMqttConf `yaml:"mqtt" json:"mqtt"`
}
type MidMqttConf struct {
	Name    string   `yaml:"name" json:"name"`
	Address []string `yaml:"address" json:"address"`
	User    string   `yaml:"user" json:"user"`
	Pwd     string   `yaml:"pwd" json:"pwd"`       //jwt 验证方式
	Scheme  string   `yaml:"scheme" json:"scheme"` //ssl | tcp | wc
}
type MidEsConf struct {
	Name                   string         `yaml:"name" json:"name"`
	Host                   []string       `yaml:"host" json:"host"`
	User                   string         `yaml:"user" json:"user"`
	Pwd                    string         `yaml:"pwd" json:"pwd"`
	AutoDetect             bool           `yaml:"auto_detect" json:"auto_detect"`
	MaxWindowResult        map[string]int `yaml:"max_window_result" json:"max_window_result"`
	TrackTotalHits         map[string]int `yaml:"track_total_hits" json:"track_total_hits"`
	DefaultMaxWindowResult int            `yaml:"default_max_window_result" json:"default_max_window_result"`
	DefaultTrackTotalHits  int            `yaml:"default_track_total_hits" json:"default_track_total_hits"`
	ReadTimeout            string         `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout           string         `yaml:"write_timeout" json:"write_timeout"`
	MaxIdleConn            string         `yaml:"max_idle_conn" json:"max_idle_conn"`
	ConnTimeout            string         `yaml:"conn_timeout" json:"conn_timeout"`
}

type ElasticConf struct {
	List []MidEsConf `yaml:"elastic" json:"elastic"`
}

func (m *MidEsConf) getMaxResult(index string) int {
	if len(m.MaxWindowResult) == 0 {
		return 0
	}
	if v, ok := m.MaxWindowResult[index]; ok {
		return v
	}
	return 0
}
