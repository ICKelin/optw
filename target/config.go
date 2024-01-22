package target

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	ListenerConfig ListenerConfig `yaml:"listener"`
	TargetConfig   TargetConfig   `yaml:"target"`
}

type ListenerConfig struct {
	Scheme     string `yaml:"scheme"`
	ListenAddr string `yaml:"listen_addr"`
	Key        string `yaml:"key"`
	Cfg        string `yaml:"raw_config"`
}

type TargetConfig struct {
	Address string `yaml:"address"`
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
