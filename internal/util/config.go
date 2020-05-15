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
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		ID       int    `yaml:"id"`
	} `yaml:"db"`
}

var Config Configuration

// init reads _config.yml into Config.
func init() {
	Config = Configuration{}
	bytes, err := ioutil.ReadFile("_config.yml")
	Chk(err)

	Chk(yaml.Unmarshal(bytes, &Config))
}
