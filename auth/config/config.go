package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	IsDebug     bool
	Env         string
	Server      Server
	Mysql       Mysql
	Redis       Redis
	ClientProxy ClientProxy
}

func (c *Config) Init() {
}

type Business struct {
}

type Server struct {
	Address string
	Port    int
	Type    string
}

type Mysql struct {
	Uri         string
	MaxConn     int
	MaxIdel     int
	MaxLifeTime int
}

type Redis struct {
	Address  string
	Password string
	Db       int
}

type ClientProxy struct {
}

func LoadConfig(fpath string) (*Config, error) {
	if len(fpath) == 0 {
		fpath = "./config.toml"
	}

	var config Config
	_, err := toml.DecodeFile(fpath, &config)
	return &config, err
}
