package app

import (
	"time"

	"github.com/visonlv/go-vkit/grpcclient"
	"github.com/visonlv/iot-engine/route/config"
	shadowpb "github.com/visonlv/iot-engine/shadow/proto"
)

type CClient struct {
	ForwardingService *shadowpb.ForwardingServiceClient
}

var (
	Cfg    *config.Config
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
		Cfg.InnerClient.Shadow,
		grpcclient.RequestTimeout(time.Second*20),
	)
	Client.ForwardingService = shadowpb.NewForwardingServiceClient("shadow", conn1)
}
