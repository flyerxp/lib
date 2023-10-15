package zookeeper

type MidConf struct {
	Name    string   `yaml:"name" json:"name"`
	Address []string `yaml:"address" json:"address"`
	User    string   `yaml:"user" json:"user"`
	Pwd     string   `yaml:"pwd" json:"pwd"`
}
type ZookeeperConf struct {
	List []MidConf `yaml:"zookeeper" json:"zookeeper"`
}
