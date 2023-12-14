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
	Mysql       Mysql
	Nats        Nats
	InnerClient InnerClient
}

type Server struct {
	Address  string
	GrpcPort int64
}

type Mysql struct {
	Uri         string
	MaxConn     int
	MaxIdel     int
	MaxLifeTime int
}

type Nats struct {
	Url      string
	Username string
	Password string
}

type InnerClient struct {
	Shadow string
	Route  string
}
