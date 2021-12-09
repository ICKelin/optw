package access

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	RouteConfig RouteConfig `yaml:"route_config"`
}

type RouteConfig struct {
	ListenerConfig ListenerConfig  `yaml:"listener"`
	NexthopConfig  []NextHopConfig `yaml:"dialer"`
}

type ListenerConfig struct {
	Scheme     string `yaml:"scheme"`
	ListenAddr string `yaml:"listen_addr"`
}

type NextHopConfig struct {
	NexthopAddr string `yaml:"nexthop_addr"`
	Scheme      string `yaml:"scheme"`
	RawConfig   string `yaml:"raw_config"`
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
