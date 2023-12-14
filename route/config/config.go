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
	return &config, err
}

type Config struct {
	Server      Server
	InnerClient InnerClient
}

type Server struct {
	Address  string
	GrpcPort int64
}

type InnerClient struct {
	Group  string
	Shadow string
}
