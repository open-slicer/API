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
	MongoDB struct {
		URI  string `yaml:"uri"`
		Name string `yaml:"name"`
	} `yaml:"mongodb"`
}

// C is the Tmpl instance of _config.yml.
var C Tmpl

// init reads _config.yml into C.
func init() {
	C = Tmpl{}
	bytes, err := ioutil.ReadFile("_config.yml")
	util.Chk(err)

	util.Chk(yaml.Unmarshal(bytes, &C))
}
