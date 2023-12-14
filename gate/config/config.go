package config

import (
	"github.com/BurntSushi/toml"
)

// InitConfig 解析配置文件
func LoadConfig(fpath string) (*Config, error) {
	if len(fpath) == 0 {
		fpath = "./config.toml"
	}
	var config Config
	_, err := toml.DecodeFile(fpath, &config)
	config.Init()
	return &config, err
}

type Config struct {
	IsDebug     bool
	Server      Server
	Business    Business
	InnerClient InnerClient
}

type Server struct {
	Address       string
	HttpPort      int32
	GrpcProxyPort int32
}

type Business struct {
	TargetList [][]string
	TargetMap  map[string]string
}

type InnerClient struct {
	Auth string
}

func (c *Config) Init() {
	c.Business.TargetMap = make(map[string]string)
	for _, v := range c.Business.TargetList {
		c.Business.TargetMap[v[0]] = v[1]
	}
}
