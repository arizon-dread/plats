package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

var (
	c    Config
	m    sync.Mutex
	once sync.Once
)

type Config struct {
	Cache Cache
	Api   []Api
}

type Cache struct {
	Url   string `yaml:"url,omitempty"`
	User  string `yaml:"user,omitempty"`
	Pass  string `yaml:"password,omitempty"`
	Port  string `yaml:"port,omitempty"`
	Proto string `yaml:"protocol,omitempty"`
}
type Api struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
	Path string `uaml:"path,omitempty"`
}

func (c *Config) Load() {
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
			if k == "Environment" {
				env = v
			}
			if k == "path" {
				path = v
			}
		}
		f, err := os.Open(fmt.Sprintf("%v/%v.yaml", path, env))
		if err != nil {
			fmt.Printf("error reading config: %v", err)
			return
		}
		var b []byte
		_, err = f.Read(b)
		if err != nil {
			fmt.Printf("unable to read from file, %v", err)
			return
		}
		err = yaml.Unmarshal(b, &c)
		if err != nil {
			fmt.Printf("unable to unmarshal config: %v", err)
			return
		}

	})
}
