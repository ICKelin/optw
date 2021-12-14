package hop

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	RouteConfig []RouteConfig `yaml:"route_config"`
}

type RouteConfig struct {
	ListenerConfig ListenerConfig `yaml:"listener"`
	HopConfig      []HopConfig    `yaml:"hops"`
}

type ListenerConfig struct {
	Scheme     string `yaml:"scheme"`
	ListenAddr string `yaml:"listen_addr"`
}

type HopConfig struct {
	HopAddr   string `yaml:"hop_addr"`
	Scheme    string `yaml:"scheme"`
	ProbeAddr string `yaml:"probe_addr"`
	RawConfig string `yaml:"raw_config"`
}

func ParseConfig(path string) (*Config, error) {
	cnt, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg = Config{}
	err = yaml.Unmarshal(cnt, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, err
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "   ")
	return string(b)
}
