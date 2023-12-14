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
	Mysql       Mysql
	TDengine    TDengine
	Group       Group
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

type Mysql struct {
	Uri         string
	MaxConn     int
	MaxIdel     int
	MaxLifeTime int
}

type TDengine struct {
	DataSource string
}

type InnerClient struct {
	Group string
	Thing string
}

type Group struct {
	Start int32
	End   int32
}
