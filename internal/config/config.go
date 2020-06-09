package config

import (
	"io/ioutil"
	"slicerapi/internal/util"

	"gopkg.in/yaml.v2"
)

// Tmpl is the main config used by the server. Read from _config.yml.
type Tmpl struct {
	HTTP struct {
		Address string `yaml:"address"`
	} `yaml:"http"`
	DB struct {
		MongoDB struct {
			URI string `yaml:"uri"`
		} `yaml:"mongodb"`
	} `yaml:"db"`
}

// Config is the Tmpl instance of _config.yml.
var Config Tmpl

// init reads _config.yml into Config.
func init() {
	Config = Tmpl{}
	bytes, err := ioutil.ReadFile("_config.yml")
	util.Chk(err)

	util.Chk(yaml.Unmarshal(bytes, &Config))
}
