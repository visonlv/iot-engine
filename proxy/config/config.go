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
	Nats        Nats
	Emqx        Emqx
	Http        Http
	Coap        Coap
	InnerClient InnerClient
}

type Server struct {
	Address  string
	GrpcPort int64
}

type Nats struct {
	Url      string
	Username string
	Password string
}

type Emqx struct {
	Url      string
	Username string
	Password string
	Timeout  int32
}

type Http struct {
	Address string
	Port    int32
}

type Coap struct {
	Address string
	Port    int32
}

type InnerClient struct {
	Group string
}
