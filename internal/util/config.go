package util

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Configuration is the main config used by the server. Read from _config.yml.
type Configuration struct {
	HTTP struct {
		Address string `yaml:"address"`
	} `yaml:"http"`
	DB struct {
		Redis struct {
			Address  string `yaml:"address"`
			Password string `yaml:"password"`
			DB       int    `yaml:"DB"`
		} `yaml:"redis"`
		Cassandra struct {
			Hosts    []string `yaml:"hosts"`
			Username string   `yaml:"username"`
			Password string   `yaml:"password"`
			Keyspace string   `yaml:"keyspace"`
		} `yaml:"cassandra"`
	} `yaml:"db"`
}

// Config is the Configuration instance of _config.yml.
var Config Configuration

// init reads _config.yml into Config.
func init() {
	Config = Configuration{}
	bytes, err := ioutil.ReadFile("_config.yml")
	Chk(err)

	Chk(yaml.Unmarshal(bytes, &Config))
}
