package app

import (
	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/iot-engine/group/config"
)

var (
	Cfg   *config.Config
	Mysql *mysqlx.MysqlClient
)

func Init(cfgPath string) {
	initConfig(cfgPath)
	initMysql()
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
