package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

var once sync.Once
var Conf *Config

type Config struct {
	Cache Cache
	Apis  []ApiHost
}

type Cache struct {
	Url   string `yaml:"url,omitempty"`
	User  string `yaml:"user,omitempty"`
	Pass  string `yaml:"password,omitempty"`
	Port  string `yaml:"port,omitempty"`
	Proto string `yaml:"protocol,omitempty"`
}
type ApiHost struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
	Path string `uaml:"path,omitempty"`
}

func Load() *Config {
	//check if config is loaded

	once.Do(func() {
		hostEnv := os.Environ()
		he := make(map[string]string)
		for _, v := range hostEnv {
			keyVal := strings.Split(v, "=")
			he[keyVal[0]] = keyVal[1]
		}
		var path, env string
		for k, v := range he {
			if k == "environment" {
				env = v
			}
			if k == "path" {
				path = v
			}
		}
		b, err := os.ReadFile(fmt.Sprintf("%v/%v.yaml", path, env))
		if err != nil {
			fmt.Printf("error reading config: %v", err)
			return
		}
		var c = &Config{}
		err = yaml.Unmarshal(b, &c)
		if err != nil {
			fmt.Printf("unable to unmarshal config: %v", err)
			return
		}
		Conf = c

	})
	return Conf
}
