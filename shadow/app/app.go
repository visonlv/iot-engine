package app

import (
	"database/sql"
	"time"

	_ "github.com/taosdata/driver-go/v3/taosRestful"
	"github.com/visonlv/go-vkit/grpcclient"
	"github.com/visonlv/go-vkit/mysqlx"
	"github.com/visonlv/iot-engine/shadow/config"
	thingpb "github.com/visonlv/iot-engine/thing/proto"
)

type CClient struct {
	ProductService *thingpb.ProductServiceClient
	RuleService    *thingpb.RuleServiceClient
}

var (
	Cfg    *config.Config
	Mysql  *mysqlx.MysqlClient
	TD     *sql.DB
	Client *CClient
)

func Init(cfgPath string) {
	initConfig(cfgPath)
	initMysql()
	initTDengine()
	initClient()
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

func initTDengine() {
	db, err := sql.Open("taosRestful", Cfg.TDengine.DataSource)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("create database if not exists iot;")
	if err != nil {
		panic(err)
	}
	TD = db
}

func initClient() {
	Client = &CClient{}
	conn1 := grpcclient.GetConnClient(
		Cfg.InnerClient.Thing,
		grpcclient.RequestTimeout(time.Second*20),
	)
	Client.ProductService = thingpb.NewProductServiceClient("thing", conn1)
	Client.RuleService = thingpb.NewRuleServiceClient("thing", conn1)
}
