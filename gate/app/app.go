package app

import (
	"time"

	"github.com/visonlv/go-vkit/grpcclient"
	authpb "github.com/visonlv/iot-engine/auth/proto"
	"github.com/visonlv/iot-engine/gate/config"
)

type CClient struct {
	AuthService *authpb.AuthServiceClient
}

var (
	Cfg    *config.Config // 配置
	Client *CClient
)

func Init(cfgPath string) {
	initConfig(cfgPath)
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

func initClient() {
	Client = &CClient{}
	conn1 := grpcclient.GetConnClient(
		Cfg.InnerClient.Auth,
		grpcclient.RequestTimeout(time.Second*20),
	)
	Client.AuthService = authpb.NewAuthServiceClient("auth", conn1)
}
