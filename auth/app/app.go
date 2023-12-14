package app

import (
	"github.com/visonlv/iot-engine/auth/config"

	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/go-vkit/redisx"
)

var (
	Cfg    *config.Config
	Mysql  *mysqlx.MysqlClient
	Client *clientProxy
	Redis  *redisx.RedisClient
)

type clientProxy struct {
}

func Init(cfgPath string) {
	initConfig(cfgPath)
	initMysql()
	initClient()
	initRedis()
}

// 初始化配置， 第一步运行
func initConfig(fpath string) {
	if cfg, err := config.LoadConfig(fpath); err == nil {
		Cfg = cfg
	} else {
		panic(err)
	}
}

func initMysql() {
	c, err := mysqlx.NewClient(Cfg.Mysql.Uri, Cfg.Mysql.MaxConn, Cfg.Mysql.MaxIdel, Cfg.Mysql.MaxLifeTime)
	if err != nil {
		panic(err)
	}
	Mysql = c
}

func initClient() {
}

func initRedis() {
	c, err := redisx.NewClient(Cfg.Redis.Address, Cfg.Redis.Password, Cfg.Redis.Db)
	if err != nil {
		panic(err)
	}
	Redis = c
}
